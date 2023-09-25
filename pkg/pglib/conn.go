package pglib

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type EnvConfig struct {
	Username string `env:"POSTGRES_USERNAME" validate:"required"`
	Password string `env:"POSTGRES_PASSWORD" validate:"required"`
	Host     string `env:"POSTGRES_HOST" validate:"required,hostname"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432" validate:"required,number"`
	Database string `env:"POSTGRES_DATABASE" envDefault:"postgres" validate:"required"`
	TimeZone string `env:"POSTGRES_TIMEZONE" envDefault:"Europe/Rome" validate:"required"`
	SSlMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable" validate:"required"`
}

type Config struct {
	Username      string
	Password      string
	Host          string
	Port          int
	Database      string
	TimeZone      string
	SSlMode       string
	SlowThreshold time.Duration
}

func FromEnvConfig(cfg EnvConfig) Config {
	return Config{
		Username: cfg.Username,
		Password: cfg.Password,
		Host:     cfg.Host,
		Port:     cfg.Port,
		Database: cfg.Database,
		TimeZone: cfg.TimeZone,
		SSlMode:  cfg.SSlMode,
	}
}

func FormatDSN(cfg Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port, cfg.SSlMode, cfg.TimeZone)
}

func Open(cfg Config) (db *gorm.DB, err error) {
	dsn := FormatDSN(cfg)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormSlog(cfg.SlowThreshold),
	})
	return
}
