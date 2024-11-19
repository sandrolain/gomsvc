package postgis

import (
	"fmt"

	"gorm.io/gorm"
)

func InRadiusScope(col string, coords string, meters int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		sql := fmt.Sprintf("ST_DWithin(%s, ST_GeomFromText(?, 4326)::geography, ?, true)", col)
		return db.Where(sql, coords, meters)
	}
}
