package repo

import "gorm.io/gorm"

func List(db *gorm.DB) {
	db.Debug().Find(nil)
}
