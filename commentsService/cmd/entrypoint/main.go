package main

import (
	"commentsService/cmd/config"
	"commentsService/pkg/database"
	"commentsService/pkg/database/models"
	"commentsService/pkg/logger"
	"commentsService/pkg/worker"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

func preload() error {
	//Инициализация конфига
	config.Init(".env")

	//Инициализация базы данных
	if _, err := database.Init(); err != nil {
		return errors.New("Ошибка при инициализации бд: " + err.Error())
	}

	//Миграции
	{
		dbQuery := fmt.Sprintf("CREATE DATABASE %s", config.Cfg.Database)
		err := database.GetDB().Exec(dbQuery).Error
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return errors.New(err.Error())
		}

		err = database.GetDB().AutoMigrate(&models.Comment{})
		if err != nil {
			return errors.New("Ошибка при миграции: " + err.Error())
		}
	}

	return nil
}

func main() {
	//Инициализируем логгер
	err := logger.New()
	if err != nil {
		log.Fatalf("%s", err.Error())
		return
	}

	//Инициализация бд + миграции
	if err := preload(); err != nil {
		logger.Error("%v", err)
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
