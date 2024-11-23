package httplib

import "github.com/gofiber/fiber/v2"

type Route struct {
	server            *Server
	ParentRoute       *Route
	Method            string
	Path              string
	Router            *fiber.Router
	validationFunc    ValidationFunc
	authorizationFunc AuthorizationFunc
}

// Valid allow to define the validation function
func (r *Route) ValidateWith(fn ValidationFunc) *Route {
	r.validationFunc = fn
	return r
}

// Auth allow to define the authorization function
func (r *Route) AuthWith(fn AuthorizationFunc) *Route {
	r.authorizationFunc = fn
	return r
}

func (s *Route) Handle(methodPath string, handler Handler) *Route {
	method, path := parsePath(methodPath)
	r := &Route{
		ParentRoute:       s,
		server:            s.server,
		Method:            method,
		Path:              path,
		validationFunc:    s.validationFunc,
		authorizationFunc: s.authorizationFunc,
	}
	router := (*s.Router).Add(r.Method, r.Path, func(ctx *fiber.Ctx) error {
		return handler(r, ctx)
	})
	r.Router = &router
	return r
}

func (r *Route) ServeStatic(path string) *Route {
	router := r.server.app.Static(r.Path, path)
	r.Router = &router
	return r
}
