package worker

import (
	"censureService/pkg/database/models"
	"censureService/pkg/logger"
	"censureService/pkg/types"
	"censureService/pkg/utilities"
	"encoding/json"
	"github.com/IBM/sarama"
	"net/http"
	"strings"
)

// KafkaWorker - структура для хранения необходимых объектов взаимодействия с Kafka
type KafkaWorker struct {
	producer            *sarama.SyncProducer
	consumer            *sarama.Consumer
	partConsumer        *sarama.PartitionConsumer
	responseTopic       string
	requestTopic        string
	requestCensureTopic string
}

// censureComment - проверяет допустим ли текст для комментария
func censureComment(str string) bool {
	if strings.Contains(str, "qwerty") || strings.Contains(str, "йцукен") || strings.Contains(str, "zxvbnm") {
		return false
	}

	return true
}

// New - создает новый экземпляр воркера с инициализацией продюсера, консьюмера и подписки на потребление запросов
func New() (*KafkaWorker, error) {
	responseTopic := "gatewayResponse"
	requestTopic := "gatewayRequest"
	requestCensureTopic := "gatewayCensureRequest"

	// Создание продюсера Kafka
	producer, err := sarama.NewSyncProducer([]string{"192.168.1.102:9092", "192.168.1.102:9093"}, nil)
	if err != nil {
		return nil, err
	}

	// Создание консьюмера Kafka
	consumer, err := sarama.NewConsumer([]string{"192.168.1.102:9092", "192.168.1.102:9093"}, nil)
	if err != nil {
		producer.Close()
		return nil, err
	}

	// Подписка на партицию "gatewayCensureRequest" в Kafka
	partConsumer, err := consumer.ConsumePartition(requestCensureTopic, 0, sarama.OffsetNewest)
	if err != nil {
		producer.Close()
		consumer.Close()
		return nil, err
	}

	kw := &KafkaWorker{
		producer:            &producer,
		consumer:            &consumer,
		partConsumer:        &partConsumer,
		responseTopic:       responseTopic,
		requestTopic:        requestTopic,
		requestCensureTopic: requestCensureTopic,
	}

	return kw, nil
}

// RunPooling - запускает горутину для прослушивание очереди запросов
func (w *KafkaWorker) RunPooling() {
	go w.handlePool()
}

// CloseKafka - закрывает продюсер, консьюмер и подписку запросы,
// вызывать в случае отказа системы или при завершении прослушивания
func (w *KafkaWorker) CloseKafka() {
	(*w.producer).Close()
	(*w.consumer).Close()
	(*w.partConsumer).Close()
}

// handlePool - просуливает очередь запросов
func (w *KafkaWorker) handlePool() {
	for {
		select {
		// (обработка входящего сообщения и отправка ответа в Kafka)
		case msg, ok := <-(*w.partConsumer).Messages():
			if !ok {
				logger.Error("Канал для прослушивания запросов закрыт!")
				w.CloseKafka()
				return
			}

			// Десериализация входящего сообщения из JSON
			var requestMessage types.Request
			logger.Info("msg: %+v", *msg)
			err := json.Unmarshal(msg.Value, &requestMessage)
			if err != nil {
				logger.Error("Ошибка при анмаршлинге JSON: %v", err)
				continue
			}

			//Ответное сообщение
			response := types.Response{ID: requestMessage.ID, ErrCode: -1}

			// Десериализация тела запроса в объект комментария
			var comment models.Comment
			err = json.Unmarshal([]byte(requestMessage.Body), &comment)
			if err != nil {
				response.ErrCode = http.StatusBadRequest
			}

			//Если запрос валидный, проверяем комментарий на цензуру
			if response.ErrCode == -1 {
				if censureComment(comment.Text) {
					response.ErrCode = http.StatusOK
				} else {
					response.ErrCode = http.StatusBadRequest
				}
			}

			//В случае невалидности запроса или комментарий нецензурный, то отправляем ошибку на gateway
			if response.ErrCode == http.StatusBadRequest {
				responseText := utilities.ToJSON(response)

				// Формируем ответное сообщение с ошибкой на gateway
				resp := &sarama.ProducerMessage{
					Topic: w.responseTopic,
					Key:   sarama.StringEncoder(response.ID),
					Value: sarama.StringEncoder(responseText),
				}
				// Отпровляем ответ в gateway
				_, _, err = (*w.producer).SendMessage(resp)
				if err != nil {
					logger.Error("Ошибка при отправке сообщения в очередь Kafka: %v", err)
				}
				continue
			}

			//Если запрос валидный и комментарий прошел цензуру, то отправляем запрос дальше в сервис комментариев
			if response.ErrCode == http.StatusOK {
				// Формируем ответное сообщение
				resp := &sarama.ProducerMessage{
					Topic: w.requestTopic,
					Key:   sarama.StringEncoder(requestMessage.ID),
					Value: sarama.StringEncoder(utilities.ToJSON(requestMessage)),
				}
				// Отпровляем ответ в gateway
				_, _, err = (*w.producer).SendMessage(resp)
				if err != nil {
					logger.Error("Ошибка при отправке сообщения в очередь Kafka: %v", err)
				}
			}
		}
	}
}
