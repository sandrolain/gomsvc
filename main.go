package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	h "github.com/sandrolain/gomscv/http"
)

func main() {
	app := h.New(h.Config{
		ValidateData: true,
		AuthorizationFunc: func(ctx *fiber.Ctx) error {
			fmt.Printf("ctx: %v\n", ctx)
			if _, ok := ctx.GetReqHeaders()["X-Token"]; !ok {
				return fmt.Errorf("Un-auth")
			}
			return nil
		},
	})

	app.FilterError(func(re *h.RouteError) *h.RouteError {
		if re.Status == 400 {
			errors := re.Error.(validator.ValidationErrors)
			fmt.Printf("errors: %v\n", errors)
			re.Body = []byte(fmt.Sprintf("Bad Request!!!! %v", re.Ctx.Request().URI()))
		}
		return re
	})

	type HelloData struct {
		XToken string `reqHeader:"X-Token"`
		Body   struct {
			Type   string `json:"ty" validate:"required,min=3,max=32"`
			Salary int    `json:"sa" validate:"required,number"`
		} `req:"body"`
		Num float64 `query:"num" validate:"min=2"`
	}

	app.With("POST /hello").Handle(h.Data(func(d *HelloData, r *h.Route, c *fiber.Ctx) error {
		fmt.Printf("d: %+v\n", d)
		c.JSON(d)
		return nil
	}))
	app.Listen(":3000")
}
