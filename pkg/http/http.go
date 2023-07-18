package http

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var reSpaces = regexp.MustCompile("\\s+")

func New(config Config, fiberConfig ...fiber.Config) *Server {
	app := fiber.New(fiberConfig...)
	server := Server{
		app:               app,
		validateData:      config.ValidateData,
		validationFunc:    config.ValidationFunc,
		authorizationFunc: config.AuthorizationFunc,
	}
	return &server
}

type Config struct {
	ValidateData      bool
	ValidationFunc    ValidationFunc
	AuthorizationFunc ValidationFunc
}

type Server struct {
	app               *fiber.App
	errorFilter       *RouteErrorFilter
	validateData      bool
	validationFunc    ValidationFunc
	authorizationFunc ValidationFunc
}

type Route struct {
	server            *Server
	Method            string
	Path              string
	Router            *fiber.Router
	validationFunc    ValidationFunc
	authorizationFunc ValidationFunc
	validate          *validator.Validate
}

type RequestData struct {
	Body interface{}
}

type Handler func(*Route, *fiber.Ctx) *RouteError

// FilterError allow to define the errors filter function
func (s *Server) FilterError(filter RouteErrorFilter) *Server {
	s.errorFilter = &filter
	return s
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func parsePath(parts ...string) (method string, path string) {
	partsNum := len(parts)
	if partsNum == 1 {
		parts = reSpaces.Split(parts[0], -1)
		partsNum = len(parts)
	}
	switch partsNum {
	case 1:
		path = parts[0]
	default:
		method = strings.ToUpper(parts[0])
		path = parts[1]
	}
	return
}

func getValidate(s *Server) (v *validator.Validate) {
	if s.validateData {
		v = validator.New()
	}
	return
}

func (s *Server) With(parts ...string) *Route {
	method, path := parsePath(parts...)
	return &Route{
		server:            s,
		Method:            method,
		Path:              path,
		validate:          getValidate(s),
		validationFunc:    s.validationFunc,
		authorizationFunc: s.authorizationFunc,
	}
}

// Valid allow to define the validation function
func (r *Route) Valid(fn ValidationFunc) *Route {
	r.validationFunc = fn
	return r
}

// Auth allow to define the authorization function
func (r *Route) Auth(fn ValidationFunc) *Route {
	r.authorizationFunc = fn
	return r
}

func (r *Route) Handle(handler Handler) *Route {
	router := r.server.app.Add(r.Method, r.Path, func(ctx *fiber.Ctx) error {
		if routeErr := handler(r, ctx); routeErr != nil {
			routeErr.Ctx = ctx
			if r.server.errorFilter != nil {
				routeErr = (*r.server.errorFilter)(routeErr)
			}
			ctx.Status(routeErr.Status)
			if len(routeErr.Body) > 0 {
				return ctx.Send(routeErr.Body)
			}
			return ctx.Send([]byte(routeErr.Error.Error()))
		}
		return nil
	})
	r.Router = &router
	return r
}

func (r *Route) Static(path string) *Route {
	router := r.server.app.Static(r.Path, path)
	r.Router = &router
	return r
}
