package user

import (
	"fmt"
	"sitex/pkg/database"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	DataBase *database.Db
}

func NewUserRepository(database *database.Db) *UserRepository {
	return &UserRepository{
		DataBase: database,
	}
}

func (repo *UserRepository) AddStatus(status statusAddInfo) error {
	// 1. Находим сотрудника
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", status.Email).First(&employee).Error; err != nil {
		return fmt.Errorf("сотрудник не найден: %w", err)
	}

	// 2. Находим ID статуса по КОДУ (а не по названию)
	var statusType StatusType
	result := repo.DataBase.DB.Where("code = ?", status.Status).First(&statusType)
	if result.Error != nil {
		return fmt.Errorf("статус с кодом '%s' не найден: %w", status.Status, result.Error)
	}

	// 3. Парсим дату
	startDate, err := time.Parse("2006-01-02", status.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты: %w", err)
	}

	// 4. Завершаем предыдущий активный статус
	if err := repo.closePreviousStatus(employee.ID, startDate); err != nil {
		return fmt.Errorf("ошибка завершения предыдущего статуса: %w", err)
	}

	// 5. Создаем новый статус
	newStatus := StatusPeriod{
		EmployeeID: employee.ID,
		StatusID:   statusType.ID, // Используем найденный ID
		StartDate:  startDate,
		EndDate:    nil,
		Comment:    status.Description,
	}

	// 6. Сохраняем
	if err := repo.DataBase.DB.Create(&newStatus).Error; err != nil {
		return fmt.Errorf("ошибка при создании статуса: %w", err)
	}

	return nil
}

// closePreviousStatus завершает предыдущий активный статус
func (repo *UserRepository) closePreviousStatus(employeeID uint, newStartDate time.Time) error {
	// Находим активный статус (без EndDate)
	var activeStatus StatusPeriod
	result := repo.DataBase.DB.
		Where("employee_id = ? AND end_date IS NULL", employeeID).
		Order("start_date DESC").
		First(&activeStatus)

	if result.Error != nil {
		// Нет активного статуса - это нормально для первого статуса
		if result.Error == gorm.ErrRecordNotFound {
			return nil
		}
		return result.Error
	}

	// Проверяем, что новая дата не раньше начала текущего статуса
	if newStartDate.Before(activeStatus.StartDate) {
		return fmt.Errorf("новая дата не может быть раньше начала текущего статуса")
	}

	// Если новая дата равна началу текущего статуса - обновляем текущий
	if newStartDate.Equal(activeStatus.StartDate) {
		return repo.DataBase.DB.
			Model(&StatusPeriod{}).
			Where("id = ?", activeStatus.ID).
			Updates(map[string]interface{}{
				"status_id": activeStatus.StatusID, // или новый statusID если нужно
				"comment":   "Обновлен",
			}).Error
	}

	// Завершаем предыдущий статус (устанавливаем EndDate за день до нового)
	endDate := newStartDate.AddDate(0, 0, -1)
	return repo.DataBase.DB.
		Model(&StatusPeriod{}).
		Where("id = ?", activeStatus.ID).
		Update("end_date", endDate).Error
}

// GetCurrentStatus возвращает текущий активный статус сотрудника
func (repo *UserRepository) GetCurrentStatus(employeeID uint) (*StatusPeriod, error) {
	var status StatusPeriod
	err := repo.DataBase.DB.
		Preload("StatusType").
		Where("employee_id = ? AND end_date IS NULL", employeeID).
		Order("start_date DESC").
		First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// GetStatusHistory возвращает историю статусов сотрудника
func (repo *UserRepository) GetStatusHistory(employeeID uint) ([]StatusPeriod, error) {
	var history []StatusPeriod
	err := repo.DataBase.DB.
		Preload("StatusType").
		Where("employee_id = ?", employeeID).
		Order("start_date DESC").
		Find(&history).Error

	return history, err
}
