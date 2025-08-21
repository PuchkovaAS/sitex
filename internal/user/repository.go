package user

import (
	"fmt"
	"sitex/internal/dt"
	"sitex/pkg/database"
	"time"
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

	// 5. Создаем новый статус
	newStatus := StatusPeriod{
		EmployeeID: employee.ID,
		StatusID:   statusType.ID, // Используем найденный ID
		StartDate:  startDate,
		Comment:    status.Description,
	}

	// 6. Сохраняем
	if err := repo.DataBase.DB.Create(&newStatus).Error; err != nil {
		return fmt.Errorf("ошибка при создании статуса: %w", err)
	}

	return nil
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

func (repo *UserRepository) GetCurrentStatus(email string) (string, error) {
	// Находим сотрудника
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", email).First(&employee).Error; err != nil {
		return "", err
	}

	// Находим последний статус с JOIN к status_types
	var statusCode string
	err := repo.DataBase.DB.
		Table("status_periods").
		Select("status_types.code").
		Joins("LEFT JOIN status_types ON status_types.id = status_periods.status_id").
		Where("status_periods.employee_id = ?", employee.ID).
		Order("status_periods.start_date DESC").
		Limit(1).
		Scan(&statusCode).Error
	if err != nil {
		return "office", nil
	}

	return statusCode, nil
}

func (repo *UserRepository) GetStatusHistory(employeeID uint) ([]StatusPeriod, error) {
	var history []StatusPeriod
	err := repo.DataBase.DB.
		Preload("StatusType").
		Where("employee_id = ?", employeeID).
		Order("start_date DESC").
		Find(&history).Error

	return history, err
}
