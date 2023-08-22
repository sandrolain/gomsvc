package api

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	slogfiber "github.com/samber/slog-fiber"
	"github.com/sandrolain/gomsvc/pkg/svc"
)

var reSpaces = regexp.MustCompile("\\s+")

func New(config Config) *Server {
	server := Server{
		validateData:      config.ValidateData,
		validationFunc:    config.ValidationFunc,
		authorizationFunc: config.AuthorizationFunc,
		errorFilter:       config.ErrorFilterFunc,
	}
	server.app = fiber.New(fiber.Config{
		ErrorHandler: getFiberErrorHandler(&server),
	})
	if config.Logger != nil {
		server.SetLogger(config.Logger)
	} else if svc.Logger() != nil {
		server.SetLogger(svc.Logger())
	}
	return &server
}

func getFiberErrorHandler(s *Server) func(ctx *fiber.Ctx, err error) error {
	return func(ctx *fiber.Ctx, err error) error {
		// Status code defaults to 500
		code := fiber.StatusInternalServerError

		// Retrieve the custom status code if it's a *fiber.Error
		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
			err = fmt.Errorf(e.Message)
		}

		// Send custom error page
		err = handleRouteError(s, &RouteError{
			Code:   fmt.Sprintf("%v", code),
			Error:  err,
			Status: code,
		}, ctx)

		if err != nil {
			// In case the SendFile fails
			return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Return from handler
		return nil
	}
}

func handleRouteError(s *Server, routeErr *RouteError, ctx *fiber.Ctx) error {
	routeErr.Ctx = ctx
	if s.errorFilter != nil {
		routeErr = s.errorFilter(routeErr)
	}
	ctx.Status(routeErr.Status)
	if len(routeErr.Body) > 0 {
		return ctx.Send(routeErr.Body)
	}
	return ctx.JSON(GetResponseForError(*routeErr))
}

type Config struct {
	Logger            *slog.Logger
	ValidateData      bool
	ValidationFunc    ValidationFunc
	AuthorizationFunc AuthorizationFunc
	ErrorFilterFunc   ErrorFilterFunc
}

type Server struct {
	app               *fiber.App
	errorFilter       ErrorFilterFunc
	validateData      bool
	validationFunc    ValidationFunc
	authorizationFunc AuthorizationFunc
}

type Route struct {
	server            *Server
	Method            string
	Path              string
	Route             *Route
	Router            *fiber.Router
	validationFunc    ValidationFunc
	authorizationFunc AuthorizationFunc
	validate          *validator.Validate
}

type RequestData struct {
	Body interface{}
}

type Handler func(*Route, *fiber.Ctx) *RouteError

// FilterError allow to define the errors filter function
func (s *Server) SetLogger(logger *slog.Logger) *Server {
	s.app.Use(slogfiber.New(logger))
	return s
}

// FilterError allow to define the errors filter function
func (s *Server) FilterError(filter ErrorFilterFunc) *Server {
	s.errorFilter = filter
	return s
}

func (s *Server) Authorize(fn AuthorizationFunc) *Server {
	s.authorizationFunc = fn
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

func (s *Server) V(version int) *Route {
	r := &Route{
		server:            s,
		validate:          getValidate(s),
		validationFunc:    s.validationFunc,
		authorizationFunc: s.authorizationFunc,
	}
	router := s.app.Group(fmt.Sprintf("/v%v", version))
	r.Router = &router
	return r
}

func (s *Server) Handle(method string, path string, handler Handler) *Route {
	r := &Route{
		server:            s,
		Method:            method,
		Path:              path,
		validate:          getValidate(s),
		validationFunc:    s.validationFunc,
		authorizationFunc: s.authorizationFunc,
	}
	router := r.server.app.Add(r.Method, r.Path, func(ctx *fiber.Ctx) error {
		if routeErr := handler(r, ctx); routeErr != nil {
			return handleRouteError(r.server, routeErr, ctx)
		}
		return nil
	})
	r.Router = &router
	return r
}

func (s *Route) Handle(method string, path string, handler Handler) *Route {
	r := &Route{
		Route:             s,
		server:            s.server,
		Method:            method,
		Path:              path,
		validate:          getValidate(s.server),
		validationFunc:    s.validationFunc,
		authorizationFunc: s.authorizationFunc,
	}
	router := (*s.Router).Add(r.Method, r.Path, func(ctx *fiber.Ctx) error {
		if routeErr := handler(r, ctx); routeErr != nil {
			return handleRouteError(s.server, routeErr, ctx)
		}
		return nil
	})
	r.Router = &router
	return r
}

// Valid allow to define the validation function
func (r *Route) Valid(fn ValidationFunc) *Route {
	r.validationFunc = fn
	return r
}

// Auth allow to define the authorization function
func (r *Route) Auth(fn AuthorizationFunc) *Route {
	r.authorizationFunc = fn
	return r
}

func (r *Route) Static(path string) *Route {
	router := r.server.app.Static(r.Path, path)
	r.Router = &router
	return r
}
