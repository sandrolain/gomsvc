package api

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

func V(version int) *Route {
	initSingleServer()
	return singleServer.V(version)
}

func ListenAddr(addr string) {
	initSingleServer()
	singleServer.Listen(addr)
}

func ListenPort(port int) {
	initSingleServer()
	singleServer.Listen(fmt.Sprintf(":%v", port))
}
