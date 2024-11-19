package dblib

import "gorm.io/gorm"

func IsQueryError(err error) bool {
	if err == nil || err == gorm.ErrRecordNotFound {
		return false
	}
	return true
}
