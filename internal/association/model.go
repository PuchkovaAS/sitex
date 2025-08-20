package association

import "gorm.io/gorm"

type UserProject struct {
	UserID    uint `gorm:"primaryKey;autoIncrement:false"`
	ProjectID uint `gorm:"primaryKey;autoIncrement:false"`
}

func (UserProject) TableName() string {
	return "user_projects"
}

func (up *UserProject) BeforeCreate(tx *gorm.DB) error {
	// Проверка на существование пользователя и проекта
	var count int64
	tx.Raw("SELECT COUNT(*) FROM users WHERE id = ?", up.UserID).Scan(&count)
	if count == 0 {
		return gorm.ErrRecordNotFound
	}

	tx.Raw("SELECT COUNT(*) FROM projects WHERE id = ?", up.ProjectID).Scan(&count)
	if count == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
