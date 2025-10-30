package resources

import (
	"errors"
	"fmt"
	"sitex/internal/user"
	"sitex/pkg/database"
	"time"

	"gorm.io/gorm"
)

type ResourceRepository struct {
	DataBase *database.Db
}

func NewRepository(database *database.Db) *ResourceRepository {
	return &ResourceRepository{
		DataBase: database,
	}
}

// AddResource добавляет новый ресурс.
// Если уже существует запись с теми же employee_id, date и status — она удаляется перед созданием новой.
func (repo *ResourceRepository) AddResource(resource resourcesAddInfo) error {
	// 1. Находим сотрудника, для которого добавляется ресурс
	var employee user.Employee
	if err := repo.DataBase.DB.Where("email = ?", resource.Email).First(&employee).Error; err != nil {
		return fmt.Errorf("сотрудник не найден по email %s: %w", resource.Email, err)
	}

	// 2. Находим сотрудника, который добавляет запись (админ)
	var whoAdded user.Employee
	if err := repo.DataBase.DB.Where("email = ?", resource.WhoAddedEmail).First(&whoAdded).Error; err != nil {
		return fmt.Errorf("сотрудник-инициатор не найден по email %s: %w", resource.WhoAddedEmail, err)
	}

	// 3. Парсим дату
	date, err := time.Parse("2006-01-02", resource.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты %s: %w", resource.Date, err)
	}

	// 5. Создаём новую запись
	newResource := Resource{
		EmployeeID:   employee.ID,
		AddedByID:    whoAdded.ID,
		Status:       resource.Status,
		Date:         date,
		ResourceName: resource.Name,
		Description:  resource.Description,
		Quantity:     resource.Quantity,
	}

	if err := repo.DataBase.DB.Create(&newResource).Error; err != nil {
		return fmt.Errorf("ошибка при создании ресурса: %w", err)
	}

	return nil
}

// DeleteResource удаляет ресурс по ID (логическое удаление).
func (repo *ResourceRepository) DeleteResource(resourceID int, emailUser, emailAdmin string) error {
	// 1. Находим сотрудника (владельца ресурса)
	var employee user.Employee
	if err := repo.DataBase.DB.Where("email = ?", emailUser).First(&employee).Error; err != nil {
		return fmt.Errorf("сотрудник не найден по email %s: %w", emailUser, err)
	}

	// 2. Находим админа, который удаляет
	var admin user.Employee
	if err := repo.DataBase.DB.Where("email = ?", emailAdmin).First(&admin).Error; err != nil {
		return fmt.Errorf("админ не найден по email %s: %w", emailAdmin, err)
	}

	// 3. Находим ресурс
	var resource Resource
	if err := repo.DataBase.DB.
		Where("id = ? AND employee_id = ?", resourceID, employee.ID).
		First(&resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("ресурс не найден или нет прав на удаление")
		}
		return fmt.Errorf("ошибка при поиске ресурса: %w", err)
	}

	// 4. Логическое удаление
	return repo.deleteResourceByID(resource.ID, admin.ID)
}

// deleteResourceByID выполняет логическое удаление ресурса.
func (repo *ResourceRepository) deleteResourceByID(resourceID uint, deletedByID uint) error {
	updates := map[string]interface{}{
		"deleted_by_id": deletedByID,
		"deleted_at":    time.Now(),
	}

	if err := repo.DataBase.DB.Model(&Resource{}).
		Where("id = ?", resourceID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("ошибка при логическом удалении ресурса: %w", err)
	}

	return nil
}

type SearchParam struct {
	Email        string
	DepartmentID uint
	SearchQuery  string
	Offset       int
	Limit        int
	Status       string
}

func (repo *ResourceRepository) getResourcesCount(searchParams SearchParam, start, end time.Time) (int64, error) {
	var count int64

	query := repo.DataBase.DB.Model(&Resource{}).
		Joins("INNER JOIN employees ON resources.employee_id = employees.id").
		Where("resources.date BETWEEN ? AND ?", start, end).
		Where("resources.deleted_at IS NULL").
		Where("employees.deleted_at IS NULL")

	if searchParams.Email != "" {
		query = query.Where("employees.email = ?", searchParams.Email)
	}
	if searchParams.DepartmentID != 0 {
		query = query.Where("employees.department_id = ?", searchParams.DepartmentID)
	}
	if searchParams.Status != "" {
		query = query.Where("resources.status = ?", searchParams.Status)
	}
	if searchParams.SearchQuery != "" {
		searchPattern := "%" + searchParams.SearchQuery + "%"
		query = query.Where(
			"employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR resources.resource_name ILIKE ? OR resources.description ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	err := query.Count(&count).Error
	return count, err
}

func (repo *ResourceRepository) GetLastResources(searchParams SearchParam) ([]Resource, int64, error) {
	var resources []Resource

	currentYear := time.Now().Year()
	startOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(currentYear, 12, 31, 23, 59, 59, 0, time.UTC)

	// Получаем общее количество
	totalCount, err := repo.getResourcesCount(searchParams, startOfYear, endOfYear)
	if err != nil {
		return nil, 0, err
	}

	// Основной запрос с Preload
	query := repo.DataBase.DB.
		Preload("Employee", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL")
		}).
		Preload("AddedBy", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL")
		}).
		Joins("INNER JOIN employees ON resources.employee_id = employees.id").
		Where("resources.date BETWEEN ? AND ?", startOfYear, endOfYear).
		Where("resources.deleted_at IS NULL").
		Where("employees.deleted_at IS NULL")

	// Применяем фильтры
	if searchParams.Email != "" {
		query = query.Where("employees.email = ?", searchParams.Email)
	}
	if searchParams.DepartmentID != 0 {
		query = query.Where("employees.department_id = ?", searchParams.DepartmentID)
	}
	if searchParams.Status != "" {
		query = query.Where("resources.status = ?", searchParams.Status)
	}
	if searchParams.SearchQuery != "" {
		searchPattern := "%" + searchParams.SearchQuery + "%"
		query = query.Where(
			"employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR resources.resource_name ILIKE ? OR resources.description ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Сортировка и пагинация
	query = query.Order("resources.updated_at DESC").
		Offset(searchParams.Offset).
		Limit(searchParams.Limit)

	// Выполняем запрос
	err = query.Find(&resources).Error
	if err != nil {
		return nil, 0, err
	}

	return resources, totalCount, nil
}
