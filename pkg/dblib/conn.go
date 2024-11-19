package dblib

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type EnvConfig struct {
	DSN      string `env:"POSTGRES_DSN" validate:"required_without_all=Username Password Host Port Database TimeZone SSLMode"`
	Username string `env:"POSTGRES_USERNAME" validate:"required_with_all=Password Host Port Database TimeZone SSLMode,required_without=DSN"`
	Password string `env:"POSTGRES_PASSWORD" validate:"required_with_all=Username Host Port Database TimeZone SSLMode,required_without=DSN"`
	Host     string `env:"POSTGRES_HOST" validate:"required_with_all=Username Password Port Database TimeZone SSLMode,required_without=DSN,omitempty,hostname"`
	Port     int    `env:"POSTGRES_PORT" validate:"required_with_all=Username Password Host Database TimeZone SSLMode,required_without=DSN,number"`
	Database string `env:"POSTGRES_DATABASE" validate:"required_with_all=Username Password Host Port TimeZone SSLMode,required_without=DSN"`
	TimeZone string `env:"POSTGRES_TIMEZONE" validate:"required_with_all=Username Password Host Port Database SSLMode,required_without=DSN"`
	SSLMode  string `env:"POSTGRES_SSLMODE" validate:"required_with_all=Username Password Host Port Database TimeZone,required_without=DSN"`
}

type Config struct {
	DSN           string `validate:"required_without_all=Username Password Host Port Database TimeZone SSLMode"`
	Username      string `validate:"required_with_all=Password Host Port Database TimeZone SSLMode,required_without=DSN"`
	Password      string `validate:"required_with_all=Username Host Port Database TimeZone SSLMode,required_without=DSN"`
	Host          string `validate:"required_with_all=Username Password Port Database TimeZone SSLMode,required_without=DSN,hostname"`
	Port          int    `validate:"required_with_all=Username Password Host Database TimeZone SSLMode,required_without=DSN,number"`
	Database      string `validate:"required_with_all=Username Password Host Port TimeZone SSLMode,required_without=DSN"`
	TimeZone      string `validate:"required_with_all=Username Password Host Port Database SSLMode,required_without=DSN"`
	SSLMode       string `validate:"required_with_all=Username Password Host Port Database TimeZone,required_without=DSN"`
	SlowThreshold time.Duration
}

func FromEnvConfig(cfg EnvConfig) Config {
	return Config{
		DSN:      cfg.DSN,
		Username: cfg.Username,
		Password: cfg.Password,
		Host:     cfg.Host,
		Port:     cfg.Port,
		Database: cfg.Database,
		TimeZone: cfg.TimeZone,
		SSLMode:  cfg.SSLMode,
	}
}

func FormatPostgresDSN(cfg Config) string {
	if cfg.DSN != "" {
		return cfg.DSN
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port, cfg.SSLMode, cfg.TimeZone)
}

func GormOpenPostgres(cfg Config) (db *gorm.DB, err error) {
	return GormOpen(
		postgres.Open(FormatPostgresDSN(cfg)),
		cfg.SlowThreshold,
	)
}

func GormOpen(dialector gorm.Dialector, slowThreshold time.Duration) (db *gorm.DB, err error) {
	db, err = gorm.Open(dialector, &gorm.Config{
		Logger: NewGormSlog(slowThreshold),
	})
	return
}
