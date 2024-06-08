package api

import (
	"apiGatewayService/pkg/journal"
	"apiGatewayService/pkg/logger"
	"github.com/IBM/sarama"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"net/http"
	"regexp"
	"sync"
	"time"
)

// jrnLog - структура для логгера
type jrnLog struct{}

// Write - реализация метода Write интерфейса io.Writer
func (l *jrnLog) Write(p []byte) (n int, err error) {
	journal.Write("%s", string(p))
	return len(p), nil
}

// CloseKafka - закрывает все что связано с Kafka
func (route Router) CloseKafka() {
	(*route.partConsumer).Close()
	(*route.consumer).Close()
	(*route.producer).Close()
}

// Router - сущность маршрутизатора, содержит приватные поля для работы исключительно внутри пакета
type Router struct {
	r                    *mux.Router
	responseChannels     map[string]chan *sarama.ConsumerMessage
	mu                   sync.Mutex
	producer             *sarama.SyncProducer
	consumer             *sarama.Consumer
	partConsumer         *sarama.PartitionConsumer
	partCensureConsumer  *sarama.PartitionConsumer
	requestTopic         string
	requestCensureTopic  string
	responseTopic        string
	responseCensureTopic string
	requestNewsTopic     string
}

// kafkaPoll - прослушивает сообщения от kafka
func (route *Router) kafkaPoll() {
	for {
		select {
		// Чтение сообщения из Kafka
		case msg, ok := <-(*route.partConsumer).Messages():
			if !ok {
				logger.Warn("Канал закрыт, выход из горутины.")
				return
			}
			responseID := string(msg.Key)
			route.mu.Lock()
			ch, exists := route.responseChannels[responseID]
			if exists {
				ch <- msg
				delete(route.responseChannels, responseID)
			}
			route.mu.Unlock()
		}
	}
}

// New - создает новый роутер для маршрутизации
func New() (*Router, error) {
	//TODO пока не придумал куда это запихнуть
	//Комментарии и цензура комментариев
	requestTopic := "gatewayRequest"
	requestCensureTopic := "gatewayCensureRequest"
	responseTopic := "gatewayResponse"
	responseCensureTopic := "gatewayCensureResponse"

	//Новости
	requestNewsTopic := "gatewayNewsRequest"

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

	// Подписка на партицию "gatewayResponse" в Kafka
	partConsumer, err := consumer.ConsumePartition(responseTopic, 0, sarama.OffsetNewest)
	if err != nil {
		producer.Close()
		consumer.Close()
		return nil, err
	}

	// Подписка на партицию "gatewayCensureResponse" в Kafka
	partCensureConsumer, err := consumer.ConsumePartition(responseCensureTopic, 0, sarama.OffsetNewest)
	if err != nil {
		producer.Close()
		consumer.Close()
		return nil, err
	}

	router := &Router{
		r:                   mux.NewRouter(),
		responseChannels:    make(map[string]chan *sarama.ConsumerMessage),
		mu:                  sync.Mutex{},
		producer:            &producer,
		consumer:            &consumer,
		partConsumer:        &partConsumer,
		requestTopic:        requestTopic,
		requestCensureTopic: requestCensureTopic,
		responseTopic:       responseTopic,
		partCensureConsumer: &partCensureConsumer,
		requestNewsTopic:    requestNewsTopic,
	}

	//Запускаем прослушку сообщений от kafka
	go router.kafkaPoll()

	return router, nil
}

// LoadEndpoints - подгружает маршруты и возвращает ссылку на экземпляр сервера
func (route *Router) LoadEndpoints(addr string) *http.Server {
	ipPortRegex := regexp.MustCompile(`^(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3}:[0-9]{1,5})|(?::[0-9]{1,5})$`)

	if !ipPortRegex.MatchString(addr) {
		addr = ":8080"
		logger.Warn("IP адрес или порт был указан неверно, установлено стандартное значение: %s", addr)
	}

	//Комментарии
	{
		route.r.HandleFunc("/comments", route.CreateComment).Methods(http.MethodPost, http.MethodOptions)
		route.r.HandleFunc("/comments/{id}", route.NewsComments).Methods(http.MethodGet, http.MethodOptions)
	}

	//Новости
	{
		route.r.HandleFunc("/news", route.News).Queries("page", "{id:[0-9]+}").Methods(http.MethodGet,
			http.MethodOptions)
		route.r.HandleFunc("/news", route.NewsByTitle).Queries("search", "title").Methods(http.MethodGet,
			http.MethodOptions)
		route.r.HandleFunc("/news", route.News).Methods(http.MethodGet,
			http.MethodOptions)
		route.r.HandleFunc("/news/{id}", route.NewsDetail).Methods(http.MethodGet, http.MethodOptions)
	}

	journalLog := &jrnLog{}

	logAfter := func(next http.Handler) http.Handler {
		return handlers.LoggingHandler(journalLog, next)
	}

	// Применяем middleware к вашему роутеру
	route.r.Use(logAfter)

	route.r.Use(cors.Default().Handler, route.middleware, mux.CORSMethodMiddleware(route.r))

	// CORS обработчик
	crs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "application/json"},
	})
	handler := crs.Handler(route.r)

	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Minute * 1,
		Handler:      cors.AllowAll().Handler(handler),
	}

	return srv
}
