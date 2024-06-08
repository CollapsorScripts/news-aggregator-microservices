package main

import (
	"censureService/pkg/logger"
	"censureService/pkg/worker"
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	//Инициализируем логгер
	err := logger.New()
	if err != nil {
		log.Fatalf("%s", err.Error())
		return
	}

	// Создание воркера Kafka
	kw, err := worker.New()
	if err != nil {
		logger.Error("%v", err)
		return
	}

	//Запускаем прослушивание очереди запросов
	kw.RunPooling()

	//Для правильного завершения приложения
	{
		wait := time.Second * 15
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		//Блокируем горутину до вызова сигнала Interrupt
		<-c

		_, cancel := context.WithTimeout(context.Background(), wait)
		defer cancel()
		kw.CloseKafka()
		os.Exit(0)
	}
}
