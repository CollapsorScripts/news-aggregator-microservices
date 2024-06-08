package main

import (
	"apiGatewayService/pkg/api"
	"apiGatewayService/pkg/journal"
	"apiGatewayService/pkg/logger"
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {

	//Инициализация логгера
	{
		if err := logger.New(); err != nil {
			log.Fatalf("Ошибка при инициализации логгера: %v", err)
		}
	}

	//Инициализация журнала
	{
		if err := journal.New(); err != nil {
			logger.Error("Ошибка при инициализации журнала: %v", err)
		}
	}

	apiRouter, err := api.New()
	if err != nil {
		logger.Error("Ошибка при создании экземпляра роутера: %v", err)
		return
	}

	defer apiRouter.CloseKafka()

	srv := apiRouter.LoadEndpoints(":8080")

	{
		wait := time.Second * 15

		// Запуск сервера в отдельном потоке
		go func() {
			logger.Info("Сервер запущен на адресе: %s", srv.Addr)
			if err := srv.ListenAndServe(); err != nil {
				log.Println(err)
			}
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		//Блокируем горутину до вызова сигнала Interrupt
		<-c

		ctx, cancel := context.WithTimeout(context.Background(), wait)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			logger.Error("%s", err.Error())
		}
		logger.Warn("Выключение сервера")
		os.Exit(0)
	}
}
