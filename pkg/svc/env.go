package svc

import (
	"github.com/caarlos0/env/v9"
	"github.com/go-playground/validator/v10"
)

func GetEnv[T any]() (cfg T, err error) {
	err = env.Parse(&cfg)
	if err != nil {
		return
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(cfg)
	if err != nil {
		return
	}
	return
}
