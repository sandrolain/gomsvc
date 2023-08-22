package api

import "github.com/gofiber/fiber/v2"

type RouteError struct {
	Error  error
	Status int
	Body   []byte
	Ctx    *fiber.Ctx
}

type ErrorFilterFunc func(*RouteError) *RouteError

type ValidationFunc func(ctx *fiber.Ctx) error
type AuthorizationFunc func(ctx *fiber.Ctx) error

func InternalServerError(ctx *fiber.Ctx, err error) *RouteError {
	return &RouteError{
		Status: fiber.StatusInternalServerError,
		Ctx:    ctx,
		Error:  err,
	}
}

func BadRequestError(ctx *fiber.Ctx, err error) *RouteError {
	return &RouteError{
		Status: fiber.StatusBadRequest,
		Ctx:    ctx,
		Error:  err,
	}
}

func ForbiddenError(ctx *fiber.Ctx, err error) *RouteError {
	return &RouteError{
		Status: fiber.StatusForbidden,
		Ctx:    ctx,
		Error:  err,
	}
}

func UnauthorizedError(ctx *fiber.Ctx, err error) *RouteError {
	return &RouteError{
		Status: fiber.StatusUnauthorized,
		Ctx:    ctx,
		Error:  err,
	}
}
