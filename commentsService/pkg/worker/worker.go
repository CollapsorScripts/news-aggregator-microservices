package worker

import (
	"commentsService/pkg/logger"
	"commentsService/pkg/types"
	"commentsService/pkg/utilities"
	"encoding/json"
	"github.com/IBM/sarama"
)

// KafkaWorker - структура для хранения необходимых объектов взаимодействия с Kafka
type KafkaWorker struct {
	producer      *sarama.SyncProducer
	consumer      *sarama.Consumer
	partConsumer  *sarama.PartitionConsumer
	responseTopic string
	requestTopic  string
}

// New - создает новый экземпляр воркера с инициализацией продюсера, консьюмера и подписки на потребление запросов
func New() (*KafkaWorker, error) {
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

	// Подписка на партицию "gatewayRequest" в Kafka
	partConsumer, err := consumer.ConsumePartition("gatewayRequest", 0, sarama.OffsetNewest)
	if err != nil {
		producer.Close()
		consumer.Close()
		return nil, err
	}

	kw := &KafkaWorker{
		producer:      &producer,
		consumer:      &consumer,
		partConsumer:  &partConsumer,
		responseTopic: "gatewayResponse",
		requestTopic:  "gatewayRequest",
	}

	//Инициализация хэндлеров
	InitHandlers()

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

			response := w.processRequest(&requestMessage)

			responseText := utilities.ToJSON(response)

			// Формируем ответное сообщение
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
		}
	}
}
