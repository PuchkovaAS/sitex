package user

import (
	"time"

	"gorm.io/gorm"
)

type statusAddForm struct {
	Status       string
	Date         string
	Description  string
	OneTimeEvent bool
}

type statusAddInfo struct {
	Email        string `gorm:"column:email"`
	Status       string `gorm:"column:status"`
	Date         string `gorm:"column:date"`
	Description  string `gorm:"column:description"`
	OneTimeEvent bool   `gorm:"column:one_time_event"`
}

type Employee struct {
	gorm.Model
	FirstName     string `gorm:"not null"`
	LastName      string `gorm:"not null"`
	Email         string `gorm:"not null;uniqueIndex:idx_employees_email"`
	PasswordHash  string `gorm:"not null"`
	Role          string `gorm:"not null;default:employee"`
	Position      string
	Department    string
	IsActive      bool           `gorm:"default:true;index:idx_employees_active"`
	StatusPeriods []StatusPeriod `gorm:"foreignKey:EmployeeID"`
}

type StatusType struct {
	gorm.Model
	Name    string         `gorm:"not null"`
	Code    string         `gorm:"not null"`
	Periods []StatusPeriod `gorm:"foreignKey:StatusID"`
}

type StatusPeriod struct {
	gorm.Model
	EmployeeID   uint      `gorm:"not null"`
	StatusID     uint      `gorm:"not null"`
	StartDate    time.Time `gorm:"not null"`
	Comment      string
	Employee     Employee   `gorm:"foreignKey:EmployeeID"`
	StatusType   StatusType `gorm:"foreignKey:StatusID"`
	OneTimeEvent bool       `gorm:"not null;default:false"`
}

type MonthHistory struct {
	Name              string
	Number            int
	WeekdayFirstMonth int
	HistoryStatus     []DayStatus
}
