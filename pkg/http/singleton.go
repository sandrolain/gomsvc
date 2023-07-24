package http

var singleServer *Server

func initSingleServer() {
	if singleServer == nil {
		singleServer = New(Config{
			ValidateData: true,
		})
	}
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

func Get[T any](path string, handler DataHandler[T]) *Route {
	return Handle("GET", path, handler)
}

func Post[T any](path string, handler DataHandler[T]) *Route {
	return Handle("POST", path, handler)
}

func Listen(addr string) {
	initSingleServer()
	singleServer.Listen(addr)
}
