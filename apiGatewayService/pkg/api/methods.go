package api

import (
	_ "apiGatewayService/pkg/journal"
	"apiGatewayService/pkg/logger"
	"apiGatewayService/pkg/utilities"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
)

// middleware - промежуточное ПО
func (route Router) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		ctx := context.WithValue(r.Context(), "id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (route Router) CreateComment(w http.ResponseWriter, r *http.Request) {
	requestID := fmt.Sprintf("%v", r.Context().Value("id"))

	dataRequest, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Ошибка при чтении тела запроса: %v", err)
		http.Error(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
	}

	request := Request{
		ID:     requestID,
		Body:   string(dataRequest),
		Path:   r.URL.Path,
		Method: r.Method,
	}

	logger.Info("Что будем отправлять: %s", utilities.ToJSON(request))

	msg := &sarama.ProducerMessage{
		Topic: route.requestCensureTopic,
		Key:   sarama.StringEncoder(requestID),
		Value: sarama.ByteEncoder(utilities.ToJSON(request)),
	}

	logger.Info("Сообщение в очередь: %+v", *msg)

	// отправка сообщения в Kafka
	_, _, err = (*route.producer).SendMessage(msg)
	if err != nil {
		logger.Error("Ошибка при отправлении сообщения в kafka: %v", err)
		http.Error(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
		return
	}

	responseCh := make(chan *sarama.ConsumerMessage)
	route.mu.Lock()
	route.responseChannels[requestID] = responseCh
	route.mu.Unlock()

	select {
	case responseMsg := <-responseCh:
		response := Response{}
		if err := json.Unmarshal(responseMsg.Value, &response); err != nil {
			logger.Error("Ошибка при десериализации ответа: %v", err)
			http.Error(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
			return
		}

		if response.ErrCode != 200 {
			logger.Error("Код ошибки: %d", response.ErrCode)
			http.Error(w, "Ошибка на стороне сервера", response.ErrCode)
			return
		}

		_, err = w.Write([]byte(response.Body))
		if err != nil {
			logger.Error("%s", err.Error())
		}
	case <-time.After(15 * time.Second):
		route.mu.Lock()
		delete(route.responseChannels, requestID)
		route.mu.Unlock()
		http.Error(w, "Таймаут ожидания ответа", http.StatusGatewayTimeout)
	}
}

func (route Router) NewsComments(w http.ResponseWriter, r *http.Request) {
	requestID := fmt.Sprintf("%v", r.Context().Value("id"))

	request := Request{
		ID:     requestID,
		Body:   "",
		Path:   r.URL.Path,
		Method: r.Method,
	}

	logger.Info("Что будем отправлять: %s", utilities.ToJSON(request))

	msg := &sarama.ProducerMessage{
		Topic: route.requestTopic,
		Key:   sarama.StringEncoder(requestID),
		Value: sarama.ByteEncoder(utilities.ToJSON(request)),
	}

	logger.Info("Сообщение в очередь: %+v", *msg)

	// отправка сообщения в Kafka
	_, _, err := (*route.producer).SendMessage(msg)
	if err != nil {
		logger.Error("Ошибка при отправлении сообщения в kafka: %v", err)
		http.Error(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
		return
	}

	responseCh := make(chan *sarama.ConsumerMessage)
	route.mu.Lock()
	route.responseChannels[requestID] = responseCh
	route.mu.Unlock()

	select {
	case responseMsg := <-responseCh:
		response := Response{}
		if err := json.Unmarshal(responseMsg.Value, &response); err != nil {
			logger.Error("Ошибка при десериализации ответа: %v", err)
			http.Error(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
			return
		}

		if response.ErrCode != 200 {
			logger.Error("Код ошибки: %d", response.ErrCode)
			http.Error(w, "Ошибка на стороне сервера", response.ErrCode)
			return
		}

		logger.Info("Ответ который пришел: %s", utilities.ToJSON(response.Body))
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(response.Body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNoContent)
			logger.Error("%s", err.Error())
		}
	case <-time.After(15 * time.Second):
		route.mu.Lock()
		delete(route.responseChannels, requestID)
		route.mu.Unlock()
		http.Error(w, "Таймаут ожидания ответа", http.StatusGatewayTimeout)
	}
}

func (route Router) News(w http.ResponseWriter, r *http.Request) {
	requestID := fmt.Sprintf("%v", r.Context().Value("id"))

	logger.Info("url: %s", (*r).URL.Path)
	url := strings.Replace(path.Join((*r).URL.Path, (*r).URL.RawQuery), "=", "/", -1)

	request := Request{
		ID:     requestID,
		Body:   "",
		Path:   url,
		Method: r.Method,
	}

	logger.Info("url: %s", url)

	msg := &sarama.ProducerMessage{
		Topic: route.requestNewsTopic,
		Key:   sarama.StringEncoder(requestID),
		Value: sarama.ByteEncoder(utilities.ToJSON(request)),
	}

	// отправка сообщения в Kafka
	_, _, err := (*route.producer).SendMessage(msg)
	if err != nil {
		logger.Error("Ошибка при отправлении сообщения в kafka: %v", err)
		http.Error(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
		return
	}

	responseCh := make(chan *sarama.ConsumerMessage)
	route.mu.Lock()
	route.responseChannels[requestID] = responseCh
	route.mu.Unlock()

	select {
	case responseMsg := <-responseCh:
		resp := Response{}

		if err := json.Unmarshal(responseMsg.Value, &resp); err != nil {
			http.Error(w, "Неверный запрос", http.StatusBadRequest)
			return
		}

		if resp.ErrCode != http.StatusOK {
			logger.Error("Неверные данные: %d", resp.ErrCode)
			http.Error(w, "Ошибка на стороне сервера", resp.ErrCode)
			return
		}

		_, err = w.Write([]byte(resp.Body))
		if err != nil {
			logger.Error("%s", err.Error())
		}
	case <-time.After(15 * time.Second):
		route.mu.Lock()
		delete(route.responseChannels, requestID)
		route.mu.Unlock()
		http.Error(w, "Таймаут ожидания ответа", http.StatusGatewayTimeout)
	}
}

func (route Router) NewsDetail(w http.ResponseWriter, r *http.Request) {
	requestID := fmt.Sprintf("%v", r.Context().Value("id"))

	url := strings.Replace(path.Join((*r).URL.Path, (*r).URL.RawQuery), "=", "/", -1)

	request := Request{
		ID:     requestID,
		Body:   "",
		Path:   url,
		Method: r.Method,
	}

	logger.Info("url: %s", url)

	msg := &sarama.ProducerMessage{
		Topic: route.requestNewsTopic,
		Key:   sarama.StringEncoder(requestID),
		Value: sarama.ByteEncoder(utilities.ToJSON(request)),
	}

	// отправка сообщения в Kafka
	_, _, err := (*route.producer).SendMessage(msg)
	if err != nil {
		logger.Error("Ошибка при отправлении сообщения в kafka: %v", err)
		http.Error(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
		return
	}

	responseCh := make(chan *sarama.ConsumerMessage)
	route.mu.Lock()
	route.responseChannels[requestID] = responseCh
	route.mu.Unlock()

	select {
	case responseMsg := <-responseCh:
		resp := Response{}

		if err := json.Unmarshal(responseMsg.Value, &resp); err != nil {
			http.Error(w, "Неверный запрос", http.StatusBadRequest)
			return
		}

		if resp.ErrCode != http.StatusOK {
			logger.Error("Неверные данные: %d", resp.ErrCode)
			http.Error(w, "Ошибка на стороне сервера", resp.ErrCode)
		}

		_, err = w.Write([]byte(resp.Body))
		if err != nil {
			logger.Error("%s", err.Error())
		}
	case <-time.After(15 * time.Second):
		route.mu.Lock()
		delete(route.responseChannels, requestID)
		route.mu.Unlock()
		http.Error(w, "Таймаут ожидания ответа", http.StatusGatewayTimeout)
	}
}

func (route Router) NewsByTitle(w http.ResponseWriter, r *http.Request) {
	requestID := fmt.Sprintf("%v", r.Context().Value("id"))

	url := (*r).URL.Path

	logger.Info("url: %s", url)

	request := Request{
		ID:     requestID,
		Body:   "",
		Path:   url,
		Method: r.Method,
	}

	logger.Info("url: %s", url)

	msg := &sarama.ProducerMessage{
		Topic: route.requestNewsTopic,
		Key:   sarama.StringEncoder(requestID),
		Value: sarama.ByteEncoder(utilities.ToJSON(request)),
	}

	// отправка сообщения в Kafka
	_, _, err := (*route.producer).SendMessage(msg)
	if err != nil {
		logger.Error("Ошибка при отправлении сообщения в kafka: %v", err)
		http.Error(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
		return
	}

	responseCh := make(chan *sarama.ConsumerMessage)
	route.mu.Lock()
	route.responseChannels[requestID] = responseCh
	route.mu.Unlock()

	select {
	case responseMsg := <-responseCh:
		resp := Response{}

		if err := json.Unmarshal(responseMsg.Value, &resp); err != nil {
			http.Error(w, "Неверный запрос", http.StatusBadRequest)
			return
		}

		if resp.ErrCode != http.StatusOK {
			logger.Error("Неверные данные: %d", resp.ErrCode)
			http.Error(w, "Ошибка на стороне сервера", resp.ErrCode)
		}

		_, err = w.Write([]byte(resp.Body))
		if err != nil {
			logger.Error("%s", err.Error())
		}
	case <-time.After(15 * time.Second):
		route.mu.Lock()
		delete(route.responseChannels, requestID)
		route.mu.Unlock()
		http.Error(w, "Таймаут ожидания ответа", http.StatusGatewayTimeout)
	}
}
