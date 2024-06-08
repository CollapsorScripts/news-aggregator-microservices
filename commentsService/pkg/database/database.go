package database

//https://gorm.io/ru_RU/docs/

import (
	"commentsService/cmd/config"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var dbase *gorm.DB

// Init - Инициализация базы данных
func Init() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", config.Cfg.DatabaseHost, config.Cfg.DatabaseUser, config.Cfg.DatabasePassword, config.Cfg.Database, config.Cfg.DatabasePort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return db, nil
}

// GetDB - Получение ссылки на экземпляр базы данных
func GetDB() *gorm.DB {
	if dbase == nil {
		dbase, _ = Init()
		sleep := time.Duration(1)
		for dbase == nil {
			sleep *= 2
			fmt.Printf("Не удалось подключиться к базе данных, повторное подключение через %d секунд", sleep)
			time.Sleep(sleep * time.Second)
			if sleep > 30 {
				sleep = 1
			}
			dbase, _ = Init()
		}
	}
	return dbase
}
