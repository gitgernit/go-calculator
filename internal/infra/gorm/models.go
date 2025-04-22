package gorm

import (
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	Login    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}

type Status int

const (
	Accepted Status = iota
	Done
)

type Expression struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	Owner  string    `gorm:"not null"`
	Status Status    `gorm:"not null"`
	Tokens []string  `gorm:"type:jsonb;not null"`
	Result float64   `gorm:"not null"`
}

var Db, _ = gorm.Open(sqlite.Open("calculator.db"), &gorm.Config{})

func Initialize() error {
	err := Db.AutoMigrate(&User{})
	err = Db.AutoMigrate(&Expression{})
	if err != nil {
		return err
	}

	return nil
}
