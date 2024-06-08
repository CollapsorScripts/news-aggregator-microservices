package types

import "newsService/pkg/database/models"

type Response struct {
	ID      string
	Body    string
	ErrCode int
}

type Request struct {
	ID     string
	Body   string
	Path   string
	Method string
}

type ResponsePagination struct {
	News        []*models.News
	CurrentPage string
	Pages       int
}
