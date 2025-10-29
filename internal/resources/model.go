package resources

import (
	"sitex/internal/user"
	"time"

	"gorm.io/gorm"
)

type resourceAddForm struct {
	Status      string
	Date        string
	Quantity    int
	Name        string
	Description string
}

type resourcesAddInfo struct {
	Status        string
	Date          string
	Quantity      int
	Name          string
	Description   string
	Email         string
	WhoAddedEmail string
}

// Resource — материальный ресурс (ноутбук, монитор и т.д.)
type Resource struct {
	gorm.Model

	// Кто владеет ресурсом (сотрудник)
	EmployeeID uint          `gorm:"not null;index"`
	Employee   user.Employee `gorm:"foreignKey:EmployeeID"`

	// Кто добавил запись
	AddedByID uint          `gorm:"not null;index"`
	AddedBy   user.Employee `gorm:"foreignKey:AddedByID"`

	// Кто удалил запись (может быть NULL, если не удалён)
	DeletedByID *uint         `gorm:"index"` // указатель — чтобы было NULL в БД
	DeletedBy   user.Employee `gorm:"foreignKey:DeletedByID"`

	// Поля из формы
	Status       string    `gorm:"not null"`           // "Учтеное", "Не учтеное"
	Date         time.Time `gorm:"not null;type:date"` // дата события (например, выдачи)
	ResourceName string    `gorm:"not null"`           // наименование
	Description  string    `gorm:"type:text"`          // описание (опционально)
	Quantity     int       `gorm:"not null;default:1"` // количество (по умолчанию 1)
}
