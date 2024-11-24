package httplib

type Routable interface {
	Handle(method string, path string, handler Handler) *Route
}

func Handle[T any](server Routable, methodPath string, handler DataReceiver[T]) *Route {
	method, path := parsePath(methodPath)
	return server.Handle(method, path, DataHandler[T](handler))
}

func Get[T any](server Routable, path string, handler DataReceiver[T]) *Route {
	return server.Handle("GET", path, DataHandler[T](handler))
}

func Post[T any](server Routable, path string, handler DataReceiver[T]) *Route {
	return server.Handle("POST", path, DataHandler[T](handler))
}

func Put[T any](server Routable, path string, handler DataReceiver[T]) *Route {
	return server.Handle("PUT", path, DataHandler[T](handler))
}

func Patch[T any](server Routable, path string, handler DataReceiver[T]) *Route {
	return server.Handle("PATCH", path, DataHandler[T](handler))
}

func Head[T any](server Routable, path string, handler DataReceiver[T]) *Route {
	return server.Handle("HEAD", path, DataHandler[T](handler))
}

func Options[T any](server Routable, path string, handler DataReceiver[T]) *Route {
	return server.Handle("OPTIONS", path, DataHandler[T](handler))
}

func Delete[T any](server Routable, path string, handler DataReceiver[T]) *Route {
	return server.Handle("DELETE", path, DataHandler[T](handler))
}
