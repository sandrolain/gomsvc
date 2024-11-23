package httplib

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type DataRequest[T any] struct {
	Data    *T
	Ctx     *fiber.Ctx
	Session *session.Session
}

func (r *DataRequest[T]) JSON(data interface{}) error {
	return r.Ctx.JSON(data)
}

func SessionValue[R any, T any](r DataRequest[T], key string) (res R, ok bool) {
	if r.Session == nil {
		return
	}
	v := r.Session.Get(key)
	res, ok = v.(R)
	return
}

type DataReceiver[T any] func(req DataRequest[T]) error

func DataHandler[T any](handler DataReceiver[T]) Handler {
	return func(r *Route, c *fiber.Ctx) error {
		var obj T
		var sess *session.Session
		var err error

		// Request authorization
		if err = authorization(r, c); err != nil {
			return err
		}

		// Request validation
		if err = validation(r, c); err != nil {
			return err
		}

		// Data load
		if err = loadData[T](c, &obj); err != nil {
			return InternalServerError(err)
		}

		// Data validation
		if err = dataValidation(r, c, &obj); err != nil {
			return err
		}

		// Handle request
		if sess, err = loadSession(r, c); err != nil {
			return err
		}

		req := DataRequest[T]{
			Ctx:     c,
			Data:    &obj,
			Session: sess,
		}
		if err = handler(req); err != nil {
			if rerr, ok := err.(RouteError); ok {
				return rerr
			}
			return InternalServerError(err)
		}

		return nil
	}
}

func loadSession(r *Route, c *fiber.Ctx) (sess *session.Session, err error) {
	if r.server.sessionStore == nil {
		return
	}
	if sess, err = r.server.sessionStore.Get(c); err != nil {
		err = InternalServerError(err)
	}
	return
}

func authorization(r *Route, c *fiber.Ctx) (err error) {
	if r.authorizationFunc == nil {
		return
	}
	if err = r.authorizationFunc(c); err != nil {
		err = UnauthorizedError(err)
	}
	return
}

func validation(r *Route, c *fiber.Ctx) (err error) {
	if r.validationFunc == nil {
		return
	}
	if err = r.validationFunc(c); err != nil {
		err = BadRequestError(err)
	}
	return
}

func dataValidation[T any](r *Route, c *fiber.Ctx, obj *T) (err error) {
	if !r.server.validateData {
		return
	}
	if err = validator.New().Struct(*obj); err != nil {
		err = BadRequestError(err)
	}
	return
}
