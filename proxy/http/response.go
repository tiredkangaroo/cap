package http

type Response struct {
	StatusCode int64
	Header     Header
	Body       *Body
}
