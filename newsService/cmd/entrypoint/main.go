package main

import (
	"context"
	"errors"
	"log"
	"newsService/cmd/config"
	"newsService/pkg/aggregator"
	"newsService/pkg/database"
	"newsService/pkg/database/models"
	"newsService/pkg/logger"
	"newsService/pkg/rss"
	"newsService/pkg/worker"
	"os"
	"os/signal"
	"time"
)

// preload - Предзагрузка конфигурации и миграция бд
func preload() error {
	//Инициализация конфига
	err := config.Init(".env")
	if err != nil {
		return err
	}

	//Инициализация базы данных
	if _, err := database.Init(); err != nil {
		return errors.New("Ошибка при инициализации бд: " + err.Error())
	}

	//Миграции
	{
		err := database.GetDB().AutoMigrate(&models.News{})
		if err != nil {
			return errors.New("Ошибка при миграции: " + err.Error())
		}
	}

	return nil
}

func main() {
	if err := logger.New(); err != nil {
		log.Fatalf("Ошибка при инициализации логгера: %v", err)
		return
	}

	if err := preload(); err != nil {
		logger.Error("%v", err)
		return
	}

	//Получаем конфигурацию
	cfgXML, err := rss.GetData()
	if err != nil {
		logger.Error("%v", err)
		return
	}

	//Создаем агрегатор
	aggr := aggregator.New(cfgXML)
	//Запускаем агрегатор
	go aggr.Start()

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
