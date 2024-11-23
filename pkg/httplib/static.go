package httplib

func Handle[T any](server *Server, methodPath string, handler DataReceiver[T]) *Route {
	return server.Handle(methodPath, DataHandler[T](handler))
}

func Get[T any](server *Server, path string, handler DataReceiver[T]) *Route {
	return server.Handle("GET "+path, DataHandler[T](handler))
}

func Post[T any](server *Server, path string, handler DataReceiver[T]) *Route {
	return server.Handle("POST "+path, DataHandler[T](handler))
}

func Put[T any](server *Server, path string, handler DataReceiver[T]) *Route {
	return server.Handle("PUT "+path, DataHandler[T](handler))
}

func Patch[T any](server *Server, path string, handler DataReceiver[T]) *Route {
	return server.Handle("PATCH "+path, DataHandler[T](handler))
}

func Head[T any](server *Server, path string, handler DataReceiver[T]) *Route {
	return server.Handle("HEAD "+path, DataHandler[T](handler))
}

func Options[T any](server *Server, path string, handler DataReceiver[T]) *Route {
	return server.Handle("OPTIONS "+path, DataHandler[T](handler))
}

func Delete[T any](server *Server, path string, handler DataReceiver[T]) *Route {
	return server.Handle("DELETE "+path, DataHandler[T](handler))
}
