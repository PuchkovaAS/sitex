package user

import (
	"time"

	"gorm.io/gorm"
)

type Employee struct {
	gorm.Model
	FirstName    string `gorm:"not null"`
	LastName     string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null;default:employee"` // employee,  admin
	Position     string
	Department   string
	IsActive     bool         `gorm:"default:true"`
	Records      []WorkRecord // Связь One-to-Many
}

type StatusType struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null"` // "В офисе", "Отпуск"
	Code string `gorm:"uniqueIndex;not null"` // "work_office", "vacation"
}

type WorkRecord struct {
	gorm.Model
	EmployeeID uint      `gorm:"not null;uniqueIndex:idx_employee_date"` // Часть составного ключа
	Date       time.Time `gorm:"not null;uniqueIndex:idx_employee_date"` // Часть составного ключа
	StatusID   uint      `gorm:"not null"`
	Comment    string

	// Связи
	Employee   Employee   `gorm:"foreignKey:EmployeeID"`
	StatusType StatusType `gorm:"foreignKey:StatusID"`
}

type StandardSchedule struct {
	gorm.Model
	Date        time.Time `gorm:"uniqueIndex;not null"`
	IsWorking   bool      `gorm:"not null"`
	Description string
}

func (Employee) TableName() string {
	return "employee"
}
