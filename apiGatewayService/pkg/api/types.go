package api

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
