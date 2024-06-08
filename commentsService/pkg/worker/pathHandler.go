package worker

import (
	"commentsService/pkg/database/models"
	"commentsService/pkg/logger"
	"commentsService/pkg/types"
	"commentsService/pkg/utilities"
	"encoding/json"
	"errors"
	"net/http"
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
	"comments": handleComment,
}

//--------------------------HANDLERS----------------------------------------

func handleComment(cmd *parsedCmd) (*types.Response, error) {
	resp := &types.Response{ID: cmd.requestID}

	switch cmd.urlMethod {
	case http.MethodPost:
		{
			var comment models.Comment
			err := json.Unmarshal([]byte(cmd.body), &comment)
			if err != nil {
				return nil, err
			}

			if err := comment.Create(); err != nil {
				return nil, err
			}
			resp.ErrCode = http.StatusOK
		}
	case http.MethodGet:
		{
			if cmd.id != 0 {
				comments, err := models.FindByNewsID(cmd.id)
				if err != nil {
					return nil, err
				}

				resp.ErrCode = http.StatusOK
				resp.Body = utilities.ToJSON(comments)
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
	var id uint = 0
	var err error = nil

	logger.Info("Parsed: %s", parsedStr[1])

	if len(parsedStr) > 1 && req.Method == http.MethodGet {
		id64, err := strconv.ParseUint(parsedStr[2], 10, 64)
		if err != nil {
			id = 0
		}

		id = uint(id64)
	}

	cmd := &parsedCmd{
		requestID: req.ID,
		method:    parsedStr[1],
		body:      req.Body,
		id:        id,
		urlMethod: req.Method,
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
