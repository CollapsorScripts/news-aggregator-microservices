package rss

import (
	"encoding/json"
	"os"
)

const test = false

type Data struct {
	URLS          []string `json:"rss"`
	RequestPeriod uint     `json:"request_period"`
}

// GetData - парсит конфигурацию и возвращает ссылку на структуру с данными
func GetData() (*Data, error) {
	data := &Data{}

	pathToConfig := "config.json"

	cfg, err := os.ReadFile(pathToConfig)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(cfg, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
