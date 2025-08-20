package user

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email      string
	Department string
	Status     string
}

func (User) TableName() string {
	return "users"
}
