package project

import (
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Name string
}

func (Project) TableName() string {
	return "projects"
}
