package dblib

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormMock struct {
	SqlDB     *sql.DB
	GormDB    *gorm.DB
	Mock      sqlmock.Sqlmock
	Error     error
	Dialector gorm.Dialector
}

func NewGormPostgresMock(slowThreshold time.Duration) (*GormMock, error) {
	os.Setenv("LOG_LEVEL", "debug")
	testDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlmock db: %w", err)
	}
	dialector := postgres.New(postgres.Config{
		Conn: testDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: NewGormSlog(slowThreshold),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm db: %w", err)
	}
	return &GormMock{
		SqlDB:     testDB,
		GormDB:    db,
		Mock:      mock,
		Error:     err,
		Dialector: dialector,
	}, nil
}
