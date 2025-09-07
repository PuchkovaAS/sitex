package user

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	gorm.Model
	Name      string     `gorm:"not null;unique;index:idx_department_name"`
	Employees []Employee `gorm:"foreignKey:DepartmentID"`
}

type ActivityInfo struct {
	StatusCount  map[string]int
	MonthHistory []MonthHistory
	CurrentMonth int
}

type MonthHistory struct {
	Name              string
	Number            int
	WeekdayFirstMonth int
	HistoryStatus     []DayStatus
}

type statusAddForm struct {
	Status       string
	Date         string
	Description  string
	OneTimeEvent bool
}

type statusAddInfo struct {
	WhoAddEmail  string `gorm:"column:who_add_email"` // Email того, кто добавляет запись
	Email        string `gorm:"column:email"`         // Email сотрудника, для которого добавляется статус
	Status       string `gorm:"column:status"`
	Date         string `gorm:"column:date"`
	Description  string `gorm:"column:description"`
	OneTimeEvent bool   `gorm:"column:one_time_event"`
}

// models/employee.go
type Employee struct {
	gorm.Model
	FirstName     string `gorm:"not null"`
	LastName      string `gorm:"not null"`
	Email         string `gorm:"not null;uniqueIndex:idx_employees_email"`
	PasswordHash  string `gorm:"not null"`
	Position      string
	DepartmentID  uint           `gorm:"index"`
	Department    Department     `gorm:"foreignKey:DepartmentID"` // Связь
	IsActive      bool           `gorm:"default:true;index:idx_employees_active"`
	IsAdmin       bool           `gorm:"default:false;index:idx_employees_admin"`
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
	WhoAddedID   uint       `gorm:"not null"`              // ID сотрудника, который добавил запись
	WhoAdded     Employee   `gorm:"foreignKey:WhoAddedID"` // Ссылка на сотрудника
	StatusType   StatusType `gorm:"foreignKey:StatusID"`
	OneTimeEvent bool       `gorm:"not null;default:false"`
}
