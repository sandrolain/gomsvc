package pglib

import "gorm.io/gorm"

func PaginateScope(page int, num int, by string, desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if num <= 0 {
			num = 10
		}
		offset := (page - 1) * num
		if desc {
			by = by + " DESC"
		}
		return db.Order(by).Offset(offset).Limit(num)
	}
}
