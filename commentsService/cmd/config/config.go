package config

import (
	"commentsService/pkg/logger"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

// Cfg - глобальный конфиг, содержащий переменные .env
var Cfg *Config

type Config struct {
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	Database         string
}

// Init - инициализация конфигурации
func Init(path string) {
	var pathToEnv string
	if len(path) > 0 {
		pathToEnv = path
	} else {
		pathToEnv = ".env"
	}

	pwd, _ := os.Getwd()
	logger.Info("path to env: %s\ndir path: %s", pathToEnv, pwd)

	err := godotenv.Load(pathToEnv)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	DatabasePort, _ := strconv.Atoi(os.Getenv("DATABASE_PORT"))

	Cfg = &Config{
		DatabaseHost:     os.Getenv("DATABASE_HOST"),
		DatabasePort:     DatabasePort,
		DatabaseUser:     os.Getenv("DATABASE_USER"),
		DatabasePassword: os.Getenv("DATABASE_PASSWORD"),
		Database:         os.Getenv("DATABASE_COMMENTS"),
	}

}
