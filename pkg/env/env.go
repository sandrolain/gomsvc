package env

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v9"
	"github.com/go-playground/validator/v10"
)

func GetEnv[T any](config *T) {
	err := env.Parse(config)
	if e, ok := err.(*env.AggregateError); ok {
		for _, er := range e.Errors {
			fmt.Fprintf(os.Stderr, "Env parse error: %v\n", er)
		}
		os.Exit(1)
	}
	v := validator.New()
	err = v.Struct(*config)
	if e, ok := err.(validator.ValidationErrors); ok {
		for _, er := range e {
			fmt.Fprintf(os.Stderr, "Env validation error: %v\n", er)
		}
		os.Exit(1)
	}
}
