package mongolib

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sandrolain/gomsvc/pkg/svc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DefaultTimeout = 2000 * time.Millisecond

type EnvClientConfig struct {
	Uri      string        `env:"MONGODB_URI" validate:"required"`
	Database string        `env:"MONGODB_DATABASE" validate:"required"`
	Timeout  time.Duration `env:"MONGODB_TIMEOUT" validate:"numeric"`
}

type ClientOptions struct {
	Uri      string `validation:"required"`
	Database string `validation:"required"`
	Timeout  time.Duration
}

func ClientOptionsFromEnvConfig(cfg EnvClientConfig) ClientOptions {
	return ClientOptions(cfg)
}

type Connection struct {
	Client      *mongo.Client
	DB          *mongo.Database
	Logger      *slog.Logger
	timeout     time.Duration
	collections map[string]*mongo.Collection
}

func Connect(config ClientOptions) (conn *Connection, err error) {
	if e := validator.New().Struct(config); e != nil {
		err = svc.Error("mongodb config not valid", e)
		return
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.Uri))
	if err != nil {
		return
	}
	conn = &Connection{
		Client:      client,
		DB:          client.Database(config.Database),
		Logger:      svc.LoggerNamespace("MongoDB"),
		timeout:     timeout,
		collections: map[string]*mongo.Collection{},
	}
	return
}
