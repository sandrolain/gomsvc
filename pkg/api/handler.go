package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type DataRequest[T any] struct {
	Data    *T
	Ctx     *fiber.Ctx
	Session *session.Session
}

type DataReceiver[T any] func(req DataRequest[T]) error

func DataHandler[T any](handler DataReceiver[T]) Handler {
	return func(r *Route, c *fiber.Ctx) error {
		var obj T

		// Request authorization
		if err := authorization(r, c); err != nil {
			return err
		}

		// Request validation
		if err := validation(r, c); err != nil {
			return err
		}

		// Data load
		if err := loadData[T](c, &obj); err != nil {
			return InternalServerError(err)
		}

		// Data validation
		if err := dataValidation(r, c, &obj); err != nil {
			return err
		}

		// Handle request
		sess, err := loadSession(c)
		if err != nil {
			return err
		}

		req := DataRequest[T]{
			Ctx:     c,
			Data:    &obj,
			Session: sess,
		}
		if err := handler(req); err != nil {
			if rerr, ok := err.(RouteError); ok {
				return rerr
			}
			return InternalServerError(err)
		}

		return nil
	}
}

func loadSession(c *fiber.Ctx) (*session.Session, error) {
	if sessionStore != nil {
		sess, e := sessionStore.Get(c)
		if e != nil {
			return nil, InternalServerError(e)
		}
		return sess, nil
	}
	return nil, nil
}

func authorization(r *Route, c *fiber.Ctx) error {
	if r.authorizationFunc != nil {
		if err := r.authorizationFunc(c); err != nil {
			return ForbiddenError(err)
		}
	}
	return nil
}

func validation(r *Route, c *fiber.Ctx) error {
	if r.validationFunc != nil {
		if err := r.validationFunc(c); err != nil {
			return BadRequestError(err)
		}
	}
	return nil
}

func dataValidation[T any](r *Route, c *fiber.Ctx, obj *T) error {
	if r.validate != nil {
		if err := r.validate.Struct(*obj); err != nil {
			return BadRequestError(err)
		}
	}
	return nil
}
