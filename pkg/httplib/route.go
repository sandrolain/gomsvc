package httplib

import "github.com/gofiber/fiber/v2"

type Route struct {
	server            *Server
	ParentRoute       *Route
	method            string
	path              string
	Router            *fiber.Router
	validationFunc    ValidationFunc
	authorizationFunc AuthorizationFunc
	noValidateData    bool
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
		ParentRoute: s,
		server:      s.server,
		method:      method,
		path:        path,
	}
	router := (*s.Router).Add(r.method, r.path, func(ctx *fiber.Ctx) error {
		return handler(r, ctx)
	})
	r.Router = &router
	return r
}

func (r *Route) Route(path string, handler ...func(*Route)) *Route {
	router := (*r.Router).Group(path)
	return &Route{
		ParentRoute: r,
		server:      r.server,
		path:        path,
		Router:      &router,
	}
}

func (r *Route) getValidationFunc() ValidationFunc {
	if r.validationFunc != nil {
		return r.validationFunc
	}
	if r.ParentRoute != nil {
		return r.ParentRoute.getValidationFunc()
	}
	return r.server.validationFunc
}

func (r *Route) getAuthorizationFunc() AuthorizationFunc {
	if r.authorizationFunc != nil {
		return r.authorizationFunc
	}
	if r.ParentRoute != nil {
		return r.ParentRoute.getAuthorizationFunc()
	}
	return r.server.authorizationFunc
}

func (r *Route) ServeStatic(path string) *Route {
	router := r.server.app.Static(r.path, path)
	r.Router = &router
	return r
}
