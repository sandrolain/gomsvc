package httplib

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	slogfiber "github.com/samber/slog-fiber"
	"github.com/sandrolain/gomsvc/pkg/certlib"
)

type ServerOptions struct {
	Logger            *slog.Logger
	ValidationFunc    ValidationFunc
	AuthorizationFunc AuthorizationFunc
	ErrorFilterFunc   ErrorFilterFunc
	TLSConfig         *certlib.ServerTLSConfigFiles `validate:"omitempty"`
}

type Server struct {
	app               *fiber.App
	errorFilter       ErrorFilterFunc
	validationFunc    ValidationFunc
	authorizationFunc AuthorizationFunc
	sessionStore      *session.Store
	tlsConfig         *tls.Config
}

func NewServer(opts ServerOptions) (res *Server, err error) {

	res = &Server{
		validationFunc:    opts.ValidationFunc,
		authorizationFunc: opts.AuthorizationFunc,
		errorFilter:       opts.ErrorFilterFunc,
	}
	res.app = fiber.New(fiber.Config{
		ErrorHandler: getFiberErrorHandler(res),
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
	})

	if opts.TLSConfig != nil {
		tlsConfig, e := certlib.LoadServerTLSConfig(*opts.TLSConfig)
		if e != nil {
			err = fmt.Errorf("failed to load credentials: %w", e)
			return
		}
		res.tlsConfig = tlsConfig
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	res.app.Use(slogfiber.New(logger))

	return
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

func (s *Server) ValidateWith(fn ValidationFunc) *Server {
	s.validationFunc = fn
	return s
}

func (s *Server) AuthWith(fn AuthorizationFunc) *Server {
	s.authorizationFunc = fn
	return s
}

func (s *Server) Handle(method string, path string, handler Handler) *Route {
	r := &Route{
		server: s,
		method: method,
		path:   path,
	}
	router := s.app.Add(r.method, r.path, func(ctx *fiber.Ctx) error {
		return handler(r, ctx)
	})
	r.Router = &router
	return r
}

func (s *Server) Route(path string, handler ...func(*Route)) (res *Route) {
	router := s.app.Group(path)
	return &Route{
		server: s,
		path:   path,
		Router: &router,
	}
}

func (s *Server) Listen(addr string, tlsConfig ...certlib.ServerTLSConfigFiles) (err error) {
	ln, e := net.Listen("tcp", addr)
	if e != nil {
		err = fmt.Errorf("failed to listen: %w", e)
		return
	}

	if s.tlsConfig != nil {
		ln = tls.NewListener(ln, s.tlsConfig)
	}

	err = s.app.Listener(ln)
	return
}
