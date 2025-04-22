package gorm

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Login    string
	Password string
}

var Db, _ = gorm.Open(sqlite.Open("calculator.db"), &gorm.Config{})

func Initialize() error {
	err := Db.AutoMigrate(&User{})
	if err != nil {
		return err
	}

	return nil
}
