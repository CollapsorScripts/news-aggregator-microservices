package aggregator

import (
	"newsService/pkg/database/models"
	"newsService/pkg/rss"
	"strings"
	"sync"
	"time"
)

// Aggregator - структура для агрегатора
type Aggregator struct {
	Response      chan []*models.News
	ErrorResponse chan error
	xmlData       *rss.Data
	m             sync.Mutex
}

// New - создание экземпляра агрегатора
func New(xmlData *rss.Data) *Aggregator {
	response := make(chan []*models.News)
	errResponse := make(chan error)
	aggr := &Aggregator{response, errResponse, xmlData, sync.Mutex{}}

	return aggr
}

// waiter - тайм аут обхода url
func waiter(dur int, done chan struct{}) {
	time.Sleep(time.Minute * time.Duration(dur))
	done <- struct{}{}
}

// handleWriteToBase - обработчик для записи в бд
// в цикле необходим continue, т.к на поле Content стоит unique, что дает возможность не дублировать новости,
// но при записи новости которая есть, получаем ошибку о дубликации, поэтому ошибку просто отправляем в канал
func (a *Aggregator) handleWriteToBase(news []*models.News) {
	for _, data := range news {
		if err := data.Create(); err != nil {
			a.ErrorResponse <- err
			continue
		}
	}
}

// handleRounder - обработчик обхода по url и вытягивание rss ленты
func (a *Aggregator) handleRounder(url string) {
	data, err := rss.Round(url)
	if err != nil {
		a.ErrorResponse <- err
		return
	}

	var news []*models.News

	for _, item := range data.Channel.Item {
		item.PubDate = strings.ReplaceAll(item.PubDate, ",", "")
		data := &models.News{
			Title:   item.Title,
			Content: item.Description,
			Link:    item.Link,
		}
		t, err := time.Parse("Mon 2 Jan 2006 15:04:05 -0700", item.PubDate)
		if err != nil {
			t, err = time.Parse("Mon 2 Jan 2006 15:04:05 GMT", item.PubDate)
		}
		if err == nil {
			data.PubTime = t.Unix()
		}
		news = append(news, data)
	}

	a.Response <- news
}

// handler - обработчик для обхода rss лент
func (a *Aggregator) handler() {
	done := make(chan struct{})
	for {
		for _, url := range a.xmlData.URLS {
			go a.handleRounder(url)
		}
		go waiter(int(a.xmlData.RequestPeriod), done)
		<-done
	}
}

// responseWorker - читает сообщения из каналов
func (a *Aggregator) responseWorker() {
	for {
		select {
		case news := <-a.Response:
			go a.handleWriteToBase(news)
			//logger.Info("Получили новость")
		case err := <-a.ErrorResponse:
			_ = err
			//logger.Error("Ошибка: ", err)
		default:
			//logger.Info("Ожидание")
		}
	}
}

// Start - запускает обработку rss ленты
func (a *Aggregator) Start() {
	go a.responseWorker()
	go a.handler()
}
