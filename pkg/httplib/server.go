package httplib

import (
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	slogfiber "github.com/samber/slog-fiber"
)

type ServerOptions struct {
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
	sessionStore      *session.Store
}

func NewServer(config ServerOptions) *Server {
	server := Server{
		validateData:      config.ValidateData,
		validationFunc:    config.ValidationFunc,
		authorizationFunc: config.AuthorizationFunc,
		errorFilter:       config.ErrorFilterFunc,
	}
	server.app = fiber.New(fiber.Config{
		ErrorHandler: getFiberErrorHandler(&server),
	})

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}
	server.app.Use(slogfiber.New(logger))

	return &server
}

func getFiberErrorHandler(s *Server) func(ctx *fiber.Ctx, err error) error {
	return func(ctx *fiber.Ctx, err error) error {
		// Status code defaults to 500

		routeErr, ok := err.(RouteError)
		if !ok {
			var code int

			// Retrieve the custom status code if it's a *fiber.Error
			e, ok := err.(*fiber.Error)
			if ok {
				code = e.Code
				if code < 400 {
					return nil
				}
			}

			if code == 0 {
				code = fiber.StatusInternalServerError
			}

			routeErr = RouteError{
				Code:   fmt.Sprintf("%v", code),
				Err:    err,
				Status: code,
			}
		}

		if s.errorFilter != nil {
			routeErr = s.errorFilter(routeErr)
		}

		ctx.Status(routeErr.Status)

		var sendErr error
		if len(routeErr.Body) > 0 {
			sendErr = ctx.Send(routeErr.Body)
		} else {
			sendErr = ctx.JSON(formatStandardResponseError(routeErr))
		}

		if sendErr != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Return from handler
		return nil
	}
}

type Handler func(*Route, *fiber.Ctx) error

// FilterError allow to define the errors filter function
func (s *Server) FilterError(filter ErrorFilterFunc) *Server {
	s.errorFilter = filter
	return s
}

func (s *Server) SessionWith(sessionProvider fiber.Storage) {
	s.sessionStore = session.New(session.Config{
		Storage: sessionProvider,
	})
}

func (s *Server) ListenAddress(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) ListenPort(port int) error {
	return s.app.Listen(fmt.Sprintf(":%v", port))
}

func (s *Server) ValidateWith(fn ValidationFunc) *Server {
	s.validationFunc = fn
	return s
}

func (s *Server) AuthWith(fn AuthorizationFunc) *Server {
	s.authorizationFunc = fn
	return s
}

func (s *Server) Handle(methodPath string, handler Handler) *Route {
	method, path := parsePath(methodPath)
	r := &Route{
		server:            s,
		Method:            method,
		Path:              path,
		validationFunc:    s.validationFunc,
		authorizationFunc: s.authorizationFunc,
	}
	router := r.server.app.Add(r.Method, r.Path, func(ctx *fiber.Ctx) error {
		return handler(r, ctx)
	})
	r.Router = &router
	return r
}
