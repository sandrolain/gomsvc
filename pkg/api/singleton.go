package api

import "log/slog"

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

func Handle[T any](method string, path string, handler DataHandler[T]) *Route {
	initSingleServer()
	return singleServer.Handle(method, path, Data(handler))
}

func V(version int) *Route {
	initSingleServer()
	return singleServer.V(version)
}

func Listen(addr string) {
	initSingleServer()
	singleServer.Listen(addr)
}
