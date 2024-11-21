package httplib

import (
	"github.com/gofiber/fiber/v2"
)

type RouteError struct {
	Err    error
	Status int
	Code   string
	Body   []byte
}

func (e RouteError) Error() string {
	return e.Err.Error()
}

type ErrorFilterFunc func(RouteError) RouteError

type ValidationFunc func(ctx *fiber.Ctx) error
type AuthorizationFunc func(ctx *fiber.Ctx) error

func Error(status int, err error) RouteError {
	return RouteError{
		Status: status,
		Err:    err,
	}
}

func InternalServerError(err error) RouteError {
	return RouteError{
		Status: fiber.StatusInternalServerError,
		Err:    err,
	}
}

func BadRequestError(err error) RouteError {
	return RouteError{
		Status: fiber.StatusBadRequest,
		Err:    err,
	}
}

func ForbiddenError(err error) RouteError {
	return RouteError{
		Status: fiber.StatusForbidden,
		Err:    err,
	}
}

func UnauthorizedError(err error) RouteError {
	return RouteError{
		Status: fiber.StatusUnauthorized,
		Err:    err,
	}
}

func UnprocessableEntityError(err error) RouteError {
	return RouteError{
		Status: fiber.StatusUnprocessableEntity,
		Err:    err,
	}
}
