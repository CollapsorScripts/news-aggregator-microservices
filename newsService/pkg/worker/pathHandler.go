package worker

import (
	"errors"
	"fmt"
	"net/http"
	"newsService/pkg/database/models"
	"newsService/pkg/logger"
	"newsService/pkg/types"
	"newsService/pkg/utilities"
	"strconv"
	"strings"
)

// parsedCmd - структура для хранения данных после парса запроса
type parsedCmd struct {
	requestID string
	method    string
	id        uint
	body      string
	urlMethod string
}

// commandHandler - типа данных для удобства работы с хэндлерами
type commandHandler func(*parsedCmd) (*types.Response, error)

// funcHandlers - глобальная мапа для хранения хэндлеров и путей к ним
var funcHandlers map[string]commandHandler

// funcHandlersBeforeInit - мапа для инициализации funcHandlers
var funcHandlersBeforeInit = map[string]commandHandler{
	"news":       handleNewsByPage,
	"newsDetail": handleDetail,
	"newsSearch": handleSearch,
}

//--------------------------HANDLERS----------------------------------------

func handleNewsByPage(cmd *parsedCmd) (*types.Response, error) {
	resp := &types.Response{ID: cmd.requestID}

	switch cmd.urlMethod {
	case http.MethodGet:
		{
			if cmd.id != 0 {
				respPg := types.ResponsePagination{}
				news, err := models.PaginationLimitWithoutContent(cmd.id)
				if err != nil {
					resp.ErrCode = http.StatusInternalServerError
				}

				pages, err := models.Paginator()
				if err != nil {
					resp.ErrCode = http.StatusInternalServerError
				}

				if resp.ErrCode != http.StatusInternalServerError {
					resp.ErrCode = http.StatusOK

					respPg.News = news
					respPg.CurrentPage = fmt.Sprintf("%d-%d", cmd.id, pages)
					respPg.Pages = pages

					resp.Body = utilities.ToJSON(respPg)
				}
				logger.Info("Ответ который будет отправлен: %s", resp.Body)

			}
		}
	}

	return resp, nil
}

func handleDetail(cmd *parsedCmd) (*types.Response, error) {
	resp := &types.Response{ID: cmd.requestID}

	logger.Info("handleDetail")

	switch cmd.urlMethod {
	case http.MethodGet:
		{
			if cmd.id != 0 {
				news := models.News{}
				if err := news.FindOne(int(cmd.id)); err != nil || news.ID == 0 {
					resp.ErrCode = http.StatusNoContent
				}

				if resp.ErrCode != http.StatusNoContent {
					resp.ErrCode = http.StatusOK

					resp.Body = utilities.ToJSON(news)
				}
				logger.Info("Ответ который будет отправлен: %s", resp.Body)
			}
		}
	}

	return resp, nil
}

func handleSearch(cmd *parsedCmd) (*types.Response, error) {
	resp := &types.Response{ID: cmd.requestID}

	logger.Info("handleSearch, cmd: %+v", *cmd)

	switch cmd.urlMethod {
	case http.MethodGet:
		{
			if cmd.id == 0 {
				news, err := models.FindByTitleWithoutContent(cmd.body)
				if err != nil {
					resp.ErrCode = http.StatusNoContent
				}

				logger.Info("Новость: %+v", news)

				if len(news) == 0 {
					resp.ErrCode = http.StatusNoContent
				}

				if resp.ErrCode != http.StatusNoContent {
					resp.ErrCode = http.StatusOK

					resp.Body = utilities.ToJSON(news)
				}
				logger.Info("Ответ который будет отправлен: %s", resp.Body)
			}
		}
	}

	return resp, nil
}

func handleUnimplemented(cmd *parsedCmd) (*types.Response, error) {
	return nil, errors.New("-3200: Данный метод не реализован")
}

//--------------------------------------------------------------------------

// funcUnimplemented - нереализованные хэндлеры
var funcUnimplemented = map[string]commandHandler{
	"unimplemented": handleUnimplemented,
}

// InitHandlers - инициализация хэндлеров
func InitHandlers() {
	funcHandlers = funcHandlersBeforeInit
}

func parseCmd(req *types.Request) (*parsedCmd, error) {
	parsedStr := strings.Split(req.Path, "/")
	method := parsedStr[1]
	urlMethod := req.Method
	body := req.Body
	var id uint = 0
	var err error = nil

	logger.Info("длина строки при поиске по id: %d", len(parsedStr))

	if req.Method == http.MethodGet {
		if len(parsedStr) == 4 {
			if parsedStr[2] == "page" {
				id64, err := strconv.ParseUint(parsedStr[3], 10, 64)
				if err != nil {
					id64 = 0
				}

				id = uint(id64)
			} else if parsedStr[2] == "search" {
				method = fmt.Sprintf("%sSearch", parsedStr[1])
				body = parsedStr[3]
			}
		}

		if len(parsedStr) == 3 && (parsedStr[2] != "page" && parsedStr[2] != "search") {
			{
				id64, err := strconv.ParseUint(parsedStr[2], 10, 64)
				if err != nil {
					id64 = 0
				}

				id = uint(id64)
				method = fmt.Sprintf("%sDetail", parsedStr[1])
			}
		}

		if len(parsedStr) == 2 && req.Path == "/news" {
			id = 1
		}

	}

	cmd := &parsedCmd{
		requestID: req.ID,
		method:    method,
		body:      body,
		id:        id,
		urlMethod: urlMethod,
	}

	return cmd, err
}

// processRequest - обработка запроса
func (w *KafkaWorker) processRequest(request *types.Request) *types.Response {
	cmd, err := parseCmd(request)
	if err != nil {
		logger.Error("Ошибка при парсе запроса: %v", err)
	}

	resp, err := standartCmdResult(cmd)
	if err != nil {
		logger.Error("Ошибка: %v", err)
		resp.ErrCode = http.StatusInternalServerError
	}

	return resp
}

func standartCmdResult(cmd *parsedCmd) (*types.Response, error) {
	handler, ok := funcHandlers[cmd.method]
	if ok {
		goto handled
	} else {
		handler = funcUnimplemented["unimplemented"]
		goto handled
	}

handled:
	return handler(cmd)
}
