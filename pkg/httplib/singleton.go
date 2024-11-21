package httplib

import (
	"fmt"
	"log/slog"
)

var singleServer *Server

func initSingleServer() {
	if singleServer == nil {
		singleServer = New(Config{
			ValidateData: true,
		})
	}
}

func SetLogger(logger *slog.Logger) {
	initSingleServer()
	singleServer.SetLogger(logger)
}

func FilterError(filter ErrorFilterFunc) {
	initSingleServer()
	singleServer.FilterError(filter)
}

func Authorize(filter AuthorizationFunc) {
	initSingleServer()
	singleServer.Authorize(filter)
}

func Handle(method string, path string, handler Handler) *Route {
	initSingleServer()
	return singleServer.Handle(method, path, handler)
}

func Get[T any](path string, handler DataReceiver[T]) *Route {
	initSingleServer()
	return singleServer.Handle("GET", path, DataHandler[T](handler))
}

func Post[T any](path string, handler DataReceiver[T]) *Route {
	initSingleServer()
	return singleServer.Handle("POST", path, DataHandler[T](handler))
}

func Put[T any](path string, handler DataReceiver[T]) *Route {
	initSingleServer()
	return singleServer.Handle("PUT", path, DataHandler[T](handler))
}

func Delete[T any](path string, handler DataReceiver[T]) *Route {
	initSingleServer()
	return singleServer.Handle("DELETE", path, DataHandler[T](handler))
}

func ListenAddr(addr string) {
	initSingleServer()
	singleServer.Listen(addr)
}

func ListenPort(port int) {
	initSingleServer()
	singleServer.Listen(fmt.Sprintf(":%v", port))
}
