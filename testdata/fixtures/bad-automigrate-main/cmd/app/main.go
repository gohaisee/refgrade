package main

import "gorm.io/gorm"

type User struct {
	ID int
}

func main() {
	var db *gorm.DB
	_ = db.AutoMigrate(&User{})
}
