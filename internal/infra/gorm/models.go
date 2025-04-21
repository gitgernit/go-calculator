package gorm

import (
	"github.com/gitgernit/go-calculator/internal/domain/auth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	auth.User
}

var Db, _ = gorm.Open(sqlite.Open("calculator.Db"), &gorm.Config{})

func Initialize() error {
	err := Db.AutoMigrate(&User{})
	if err != nil {
		return err
	}

	return nil
}
