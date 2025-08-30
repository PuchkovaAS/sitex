package user

import (
	"errors"
	"fmt"
	"sitex/internal/dt"
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
	// 1. Находим сотрудника, для которого добавляется статус
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", status.Email).First(&employee).Error; err != nil {
		return fmt.Errorf("сотрудник не найден: %w", err)
	}

	// 2. Находим сотрудника, который добавляет запись
	var whoAdded Employee
	if err := repo.DataBase.DB.Where("email = ?", status.WhoAddEmail).First(&whoAdded).Error; err != nil {
		return fmt.Errorf("сотрудник, добавляющий запись, не найден: %w", err)
	}

	// 3. Находим ID статуса по КОДУ
	var statusType StatusType
	result := repo.DataBase.DB.Where("code = ?", status.Status).First(&statusType)
	if result.Error != nil {
		return fmt.Errorf("статус с кодом '%s' не найден: %w", status.Status, result.Error)
	}

	// 4. Парсим дату
	startDate, err := time.Parse("2006-01-02", status.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты: %w", err)
	}

	// 5. Проверяем, есть ли уже запись на эту дату
	var existingRecord StatusPeriod
	result = repo.DataBase.DB.
		Where("employee_id = ? AND start_date = ?", employee.ID, startDate.Format("2006-01-02")).
		First(&existingRecord)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Записи нет - СОЗДАЕМ новую
			newStatus := StatusPeriod{
				EmployeeID:   employee.ID,
				StatusID:     statusType.ID,
				StartDate:    startDate,
				OneTimeEvent: status.OneTimeEvent,
				Comment:      status.Description,
				WhoAddedID:   whoAdded.ID, // Добавляем информацию о том, кто создал запись
			}
			return repo.DataBase.DB.Create(&newStatus).Error
		}
		return fmt.Errorf("ошибка при поиске существующей записи: %w", result.Error)
	}

	// Запись существует - ОБНОВЛЯЕМ
	return repo.DataBase.DB.
		Model(&StatusPeriod{}).
		Where("id = ?", existingRecord.ID).
		Updates(map[string]any{
			"status_id":      statusType.ID,
			"one_time_event": status.OneTimeEvent,
			"comment":        status.Description,
			"who_added_id":   whoAdded.ID, // Обновляем информацию о том, кто изменил запись
			"updated_at":     time.Now(),
		}).Error
}

func (repo *UserRepository) GetUserInfo(email string) (dt.UserInfo, error) {
	var user Employee

	// Получаем только нужные поля
	err := repo.DataBase.DB.
		Where("email = ?", email).
		Select("first_name, last_name, role, position").
		First(&user).Error
	if err != nil {
		return dt.UserInfo{}, err
	}

	// Создаем структуру с нужными полями
	return dt.UserInfo{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Position:  user.Position,
	}, nil
}

func (repo *UserRepository) GetLastStatus(email string, date time.Time) (string, error) {
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", email).First(&employee).Error; err != nil {
		return "", err
	}

	var statusName string

	err := repo.DataBase.DB.
		Table("status_periods").
		Select("status_types.name").
		Joins("LEFT JOIN status_types ON status_types.id = status_periods.status_id").
		Where("status_periods.employee_id = ?", employee.ID).
		Where("status_periods.start_date <= ?", date).
		Where("status_periods.one_time_event = ?", false).
		Order("status_periods.start_date DESC").
		Limit(1).
		Scan(&statusName).Error
	if err != nil {
		return "В офисе", nil
	}

	return statusName, nil
}

func (repo *UserRepository) GetCurrentStatus(email string, date time.Time) (string, error) {
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", email).First(&employee).Error; err != nil {
		return "", err
	}

	var statusName string

	err := repo.DataBase.DB.
		Table("status_periods").
		Select("status_types.name").
		Joins("LEFT JOIN status_types ON status_types.id = status_periods.status_id").
		Where("status_periods.employee_id = ?", employee.ID).
		Where("status_periods.start_date = ?", date).
		Order("status_periods.start_date DESC").
		Limit(1).
		Scan(&statusName).Error

	if err == nil && statusName != "" {
		return statusName, nil
	}

	statusName, err = repo.GetLastStatus(email, date)
	return statusName, err
}

func (repo *UserRepository) DeleteStatus(statusID int, email string) error {
	// Находим сотрудника
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", email).First(&employee).Error; err != nil {
		return fmt.Errorf("сотрудник не найден: %w", err)
	}

	// Удаляем статус с проверкой принадлежности сотруднику
	result := repo.DataBase.DB.
		Where("id = ? AND employee_id = ?", statusID, employee.ID).
		Delete(&StatusPeriod{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("статус не найден или нет прав для удаления")
	}

	return nil
}

func (repo *UserRepository) GetLastAddStatus(email string, limit ...int) ([]StatusPeriod, error) {
	var history []StatusPeriod

	// Получаем текущий год
	currentYear := time.Now().Year()
	startOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(currentYear, 12, 31, 23, 59, 59, 0, time.UTC)

	query := repo.DataBase.DB.
		Preload("Employee").
		Preload("WhoAdded").
		Preload("StatusType").
		Joins("INNER JOIN employees ON status_periods.employee_id = employees.id").
		Where("employees.email = ?", email).
		Where("status_periods.start_date BETWEEN ? AND ?", startOfYear, endOfYear).
		Order("status_periods.updated_at DESC")

	// Если передан лимит, добавляем его
	if len(limit) > 0 && limit[0] > 0 {
		query = query.Limit(limit[0])
	}

	err := query.Find(&history).Error
	return history, err
}

func (repo *UserRepository) GetStatusHistory(
	email string,
	timeStart, timeEnd time.Time,
) ([]StatusPeriod, error) {
	var history []StatusPeriod

	err := repo.DataBase.DB.
		Preload("Employee").
		Preload("StatusType"). // Загружаем связанный StatusType
		Joins("INNER JOIN employees ON status_periods.employee_id = employees.id").
		Where("employees.email = ?", email).
		Where("start_date >= ?", timeStart).
		Where("start_date <= ?", timeEnd).
		Order("start_date DESC").
		Find(&history).
		Error

	return history, err
}
