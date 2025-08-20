package user

import (
	"time"

	"gorm.io/gorm"
)

type statusAddForm struct {
	Status      string
	Date        string
	Description string
}

type statusAddInfo struct {
	Email       string
	Status      string
	Date        string
	Description string
}

type Employee struct {
	gorm.Model
	FirstName     string `gorm:"not null"`
	LastName      string `gorm:"not null"`
	Email         string `gorm:"not null"`
	PasswordHash  string `gorm:"not null"`
	Role          string `gorm:"not null;default:employee"`
	Position      string
	Department    string
	IsActive      bool `gorm:"default:true"`
	StatusPeriods []StatusPeriod
}

type StatusType struct {
	gorm.Model
	Name string `gorm:"not null"`
	Code string `gorm:"not null"`
}

type StatusPeriod struct {
	gorm.Model
	EmployeeID uint      `gorm:"not null"`
	StatusID   uint      `gorm:"not null"`
	StartDate  time.Time `gorm:"not null"`
	EndDate    *time.Time
	Comment    string
}

type OfficialHoliday struct {
	gorm.Model
	Date         time.Time `gorm:"not null"`
	Name         string    `gorm:"not null"`
	Type         string    `gorm:"not null;size:20"`
	Description  string
	OriginalDate *time.Time
}
