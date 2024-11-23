package devlib

import (
	"log/slog"

	"github.com/k0kubun/pp/v3"
)

func P(v any) {
	_, err := pp.Print(v)
	if err != nil {
		slog.Default().Error(err.Error())
	}
}
