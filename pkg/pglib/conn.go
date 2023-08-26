package pglib

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Dsn           string
	SlowThreshold time.Duration
}

func Open(cfg Config) (db *gorm.DB, err error) {
	db, err = gorm.Open(postgres.Open(cfg.Dsn), &gorm.Config{
		Logger: NewGormSlog(cfg.SlowThreshold),
	})
	return
}
