package httplib

import (
	"fmt"

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

type ValidationFunc func(ctx *fiber.Ctx, r *Route) error
type AuthorizationFunc func(ctx *fiber.Ctx, r *Route) error

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

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ResponseErrorEnvelope struct {
	Error ResponseError `json:"error,omitempty"`
}

func formatStandardResponseError(err RouteError) ResponseErrorEnvelope {
	code := err.Code
	if code == "" {
		code = fmt.Sprintf("%v", err.Status)
	}
	return ResponseErrorEnvelope{
		Error: ResponseError{
			Code:    code,
			Message: err.Error(),
		},
	}
}
