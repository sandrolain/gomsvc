package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type DataHandler[T any] func(data *T, ctx *fiber.Ctx) error

func Data[T any](handler DataHandler[T]) Handler {
	return func(r *Route, c *fiber.Ctx) *RouteError {
		var obj T

		fmt.Printf("c.Get(\"content-length\"): %v\n", c.Get("content-length"))

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
			return InternalServerError(c, err)
		}

		// Data validation
		if err := dataValidation(r, c, &obj); err != nil {
			return err
		}

		// Handle request
		if err := handler(&obj, c); err != nil {
			return InternalServerError(c, err)
		}

		return nil
	}
}

func authorization(r *Route, c *fiber.Ctx) *RouteError {
	if r.authorizationFunc != nil {
		if err := r.authorizationFunc(c); err != nil {
			return ForbiddenError(c, err)
		}
	}
	return nil
}

func validation(r *Route, c *fiber.Ctx) *RouteError {
	if r.validationFunc != nil {
		if err := r.validationFunc(c); err != nil {
			return BadRequestError(c, err)
		}
	}
	return nil
}

func dataValidation[T any](r *Route, c *fiber.Ctx, obj *T) *RouteError {
	if r.validate != nil {
		if err := r.validate.Struct(*obj); err != nil {
			return BadRequestError(c, err)
		}
	}
	return nil
}
