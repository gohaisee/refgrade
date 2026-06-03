package repo

import "gorm.io/gorm"

func FindByName(db *gorm.DB, name string) {
	db.Raw("SELECT * FROM users WHERE name = " + name).Scan(nil)
}
