package rss

import (
	"encoding/xml"
	"io"
	"net/http"
	"newsService/pkg/logger"
)

type item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type items struct {
	Item []item `xml:"item"`
}

// XML - структура новостей
type XML struct {
	Channel items `xml:"channel"`
}

// Round - обрабатывает xml страницу и возвращает структуру с данными
func Round(url string) (*XML, error) {
	news := &XML{}
	resp, err := http.Get(url)
	if err != nil {
		logger.Error("Ошибка при get: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Чтение тела ответа
	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Ошибка при чтении данных: %v", err)
		return nil, err
	}
	err = xml.Unmarshal(xmlData, &news)
	if err != nil {
		logger.Error("Ошибка при десерелиазации: %v", err)
		return nil, err
	}

	return news, nil
}
