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

func (repo *UserRepository) ChangePassword(email string, password string) error {
	var user Employee
	result := repo.DataBase.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("пользователь с email %s не найден", email)
		}
		return fmt.Errorf("ошибка базы данных: %w", result.Error)
	}
	if user.ID == 0 {
		return fmt.Errorf("пользователь не найден")
	}
	// Обновление пароля
	result = repo.DataBase.DB.Model(&user).
		Update("password_hash", password)

	if result.Error != nil {
		return fmt.Errorf("ошибка при обновлении пароля: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("пароль не был изменен")
	}

	return nil
}

func (repo *UserRepository) GetAllDepartments() ([]Department, error) {
	var departments []Department
	err := repo.DataBase.Find(&departments).Error
	return departments, err
}

func (repo *UserRepository) FindOrCreateDepartment(departmentName string) (Department, error) {
	// Найти или создать отдел
	var department Department
	result := repo.DataBase.Where("name = ?", departmentName).First(&department)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Создать новый отдел
			department = Department{Name: departmentName}
			if err := repo.DataBase.Create(&department).Error; err != nil {
				return Department{}, err
			}
		} else {
			return Department{}, result.Error
		}
	}

	return department, nil
}

func (repo *UserRepository) CreateUserWithDepartment(
	FirstName, LastName, Email, DepartmentName, Position, Password string,
	IsActive, IsAdmin, ShowTimeEvents bool,
) error {
	user := &Employee{
		FirstName:      FirstName,
		LastName:       LastName,
		Position:       Position,
		IsAdmin:        IsAdmin,
		IsActive:       IsActive,
		PasswordHash:   Password,
		Email:          Email,
		ShowTimeEvents: ShowTimeEvents,
	}
	department, err := repo.FindOrCreateDepartment(DepartmentName)
	if err != nil {
		return err
	}

	user.DepartmentID = department.ID
	return repo.DataBase.Create(user).Error
}

func (repo *UserRepository) EmailExists(email string) bool {
	var count int64
	result := repo.DataBase.Model(&Employee{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false
	}
	return count > 0
}

func (repo *UserRepository) GetEmployeeInfo(email string) (Employee, error) {
	var employee Employee
	if err := repo.DataBase.DB.
		Preload("Department").
		Where("email = ?", email).First(&employee).Error; err != nil {
		return Employee{}, fmt.Errorf("сотрудник не найден: %w", err)
	}
	return employee, nil
}

func (repo *UserRepository) GetEmployeeName(email string) (string, error) {
	var employee Employee
	if err := repo.DataBase.DB.
		Preload("Department").
		Where("email = ?", email).First(&employee).Error; err != nil {
		return "", fmt.Errorf("сотрудник не найден: %w", err)
	}
	return employee.LastName + " " + employee.FirstName, nil
}

func (repo *UserRepository) AddTimeEvent(event timeEventAddInfo) error {
	// 1. Находим сотрудника, для которого добавляется событие
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", event.Email).First(&employee).Error; err != nil {
		return fmt.Errorf("сотрудник не найден: %w", err)
	}

	// 2. Находим сотрудника, который добавляет запись
	var whoAdded Employee
	if err := repo.DataBase.DB.Where("email = ?", event.WhoAddEmail).First(&whoAdded).Error; err != nil {
		return fmt.Errorf("сотрудник, добавляющий запись, не найден: %w", err)
	}

	// 3. Находим тип события по коду
	var eventType TimeEventType
	if err := repo.DataBase.DB.Where("code = ?", event.EventType).First(&eventType).Error; err != nil {
		return fmt.Errorf("тип события с кодом '%s' не найден: %w", event.EventType, err)
	}

	// 4. Парсим дату
	eventDate, err := time.Parse("2006-01-02", event.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты: %w", err)
	}

	// 5. Парсим время
	scheduled, err := time.Parse("15:04", event.ScheduledTime)
	if err != nil {
		return fmt.Errorf("неверный формат планового времени: %w", err)
	}
	actual, err := time.Parse("15:04", event.ActualTime)
	if err != nil {
		return fmt.Errorf("неверный формат фактического времени: %w", err)
	}

	// 6. Рассчитываем разницу в минутах
	var diffMinutes int
	switch eventType.Code {
	case "late":
		diffMinutes = int(actual.Sub(scheduled).Minutes())
		if diffMinutes <= 0 {
			return fmt.Errorf("для опоздания фактическое время должно быть позже планового")
		}
	case "early_leave":
		diffMinutes = int(scheduled.Sub(actual).Minutes())
		if diffMinutes <= 0 {
			return fmt.Errorf("для раннего ухода фактическое время должно быть раньше планового")
		}
	}

	// 7. Проверяем, существует ли уже событие с такими employee_id, date, event_type_id
	var existingEvent TimeEvent
	result := repo.DataBase.DB.
		Where("employee_id = ? AND date = ? AND event_type_id = ?", employee.ID, eventDate, eventType.ID).
		First(&existingEvent)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Записи нет — создаём новую
			newEvent := TimeEvent{
				EmployeeID:    employee.ID,
				WhoAddedID:    whoAdded.ID,
				EventTypeID:   eventType.ID,
				Date:          eventDate,
				ScheduledTime: event.ScheduledTime,
				ActualTime:    event.ActualTime,
				Description:   event.Description,
				DifferenceMin: diffMinutes,
			}
			return repo.DataBase.DB.Create(&newEvent).Error
		}
		// Другая ошибка (например, проблемы с БД)
		return fmt.Errorf("ошибка при поиске существующего события: %w", result.Error)
	}

	repo.DeleteTimeEvent(int(existingEvent.ID), event.Email, event.WhoAddEmail)

	newEvent := TimeEvent{
		EmployeeID:    employee.ID,
		WhoAddedID:    whoAdded.ID,
		EventTypeID:   eventType.ID,
		Date:          eventDate,
		ScheduledTime: event.ScheduledTime,
		ActualTime:    event.ActualTime,
		Description:   event.Description,
		DifferenceMin: diffMinutes,
	}
	return repo.DataBase.DB.Create(&newEvent).Error
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

	// Получаем пользователя с предзагрузкой отдела
	err := repo.DataBase.DB.
		Preload("Department").
		Where("email = ?", email).
		Select("first_name, last_name, email, position, is_admin, is_active, department_id, show_time_events").
		First(&user).Error
	if err != nil {
		return dt.UserInfo{}, err
	}

	// Создаем UserInfo
	userInfo := dt.UserInfo{
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Position:       user.Position,
		IsAdmin:        user.IsAdmin,
		IsActive:       user.IsActive,
		ShowTimeEvents: user.ShowTimeEvents,
	}

	// Проверяем, загружен ли отдел (DepartmentID != 0 и Department.Name не пустое)
	if user.DepartmentID != 0 && user.Department.Name != "" {
		userInfo.Department = user.Department.Name
	}

	return userInfo, nil
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
		Where("status_periods.deleted_at IS NULL"). // Исключаем удаленные статусы
		Where("status_types.deleted_at IS NULL").   // Также исключаем удаленные статус-типы
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
		Unscoped(). // Отключаем soft delete по умолчанию
		Table("status_periods").
		Select("status_types.name").
		Joins("LEFT JOIN status_types ON status_types.id = status_periods.status_id").
		Where("status_periods.employee_id = ?", employee.ID).
		Where("status_periods.start_date = ?", date).
		Where("status_periods.deleted_at IS NULL"). // Явно исключаем удаленные
		Where("status_types.deleted_at IS NULL").   // Явно исключаем удаленные
		Order("status_periods.start_date DESC").
		Limit(1).
		Scan(&statusName).Error

	if err == nil && statusName != "" {
		return statusName, nil
	}

	statusName, err = repo.GetLastStatus(email, date)
	return statusName, err
}

func (repo *UserRepository) DeleteTimeEvent(timeEventID int, emailUser, emailAdmin string) error {
	// 1. Находим сотрудника по email
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", emailUser).First(&employee).Error; err != nil {
		return fmt.Errorf("сотрудник не найден: %w", err)
	}

	var employeeAdmin Employee
	if err := repo.DataBase.DB.Where("email = ?", emailAdmin).First(&employeeAdmin).Error; err != nil {
		return fmt.Errorf("сотрудник не найден: %w", err)
	}

	// 2. Находим событие, чтобы проверить права и обновить who_deleted_id
	var event TimeEvent
	if err := repo.DataBase.DB.
		Where("id = ? AND employee_id = ?", timeEventID, employee.ID).
		First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("событие не найдено или нет прав для удаления")
		}
		return fmt.Errorf("ошибка при поиске события: %w", err)
	}

	// 3. Обновляем who_deleted_id и помечаем как удалённое (soft delete)
	result := repo.DataBase.DB.Model(&event).
		Updates(map[string]interface{}{
			"who_deleted_id": employeeAdmin.ID,
			"deleted_at":     time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении события: %w", result.Error)
	}

	return nil
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
		return fmt.Errorf("ошибка при удалении события: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("статус не найден или нет прав для удаления")
	}

	return nil
}

// Отдельная функция для подсчета общего количества статус-периодов
func (repo *UserRepository) getStatusPeriodsCount(searchParams SearchParam, startOfYear, endOfYear time.Time) (int64, error) {
	var totalCount int64

	query := repo.DataBase.DB.
		Model(&StatusPeriod{}).
		Joins("INNER JOIN employees ON status_periods.employee_id = employees.id").
		Where("status_periods.start_date BETWEEN ? AND ?", startOfYear, endOfYear).
		Where("status_periods.deleted_at IS NULL").
		Where("employees.deleted_at IS NULL")

	// Применяем фильтры
	if searchParams.Email != "" {
		query = query.Where("employees.email = ?", searchParams.Email)
	}
	if searchParams.DepartmentID != 0 {
		query = query.Where("employees.department_id = ?", searchParams.DepartmentID)
	}
	if searchParams.SearchQuery != "" {
		searchPattern := "%" + searchParams.SearchQuery + "%"
		query = query.Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ?",
			searchPattern, searchPattern)
	}

	err := query.Count(&totalCount).Error
	return totalCount, err
}

func (repo *UserRepository) getTimeEventsCount(searchParams SearchParam, start, end time.Time) (int64, error) {
	var count int64

	query := repo.DataBase.DB.Model(&TimeEvent{}).
		Joins("INNER JOIN employees ON time_events.employee_id = employees.id").
		Where("time_events.date BETWEEN ? AND ?", start, end).
		Where("time_events.deleted_at IS NULL").
		Where("employees.deleted_at IS NULL")

	if searchParams.Email != "" {
		query = query.Where("employees.email = ?", searchParams.Email)
	}
	if searchParams.DepartmentID != 0 {
		query = query.Where("employees.department_id = ?", searchParams.DepartmentID)
	}
	if searchParams.SearchQuery != "" {
		searchPattern := "%" + searchParams.SearchQuery + "%"
		query = query.Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ?",
			searchPattern, searchPattern)
	}

	err := query.Count(&count).Error
	return count, err
}

func (repo *UserRepository) GetLastAddTimeEvents(searchParams SearchParam) ([]TimeEvent, int64, error) {
	var events []TimeEvent

	currentYear := time.Now().Year()
	startOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(currentYear, 12, 31, 23, 59, 59, 0, time.UTC)

	// Получаем общее количество
	totalCount, err := repo.getTimeEventsCount(searchParams, startOfYear, endOfYear)
	if err != nil {
		return nil, 0, err
	}

	// Основной запрос с Preload
	query := repo.DataBase.DB.
		Preload("Employee", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL")
		}).
		Preload("WhoAdded", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL")
		}).
		Preload("EventType", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL")
		}).
		Joins("INNER JOIN employees ON time_events.employee_id = employees.id").
		Where("time_events.date BETWEEN ? AND ?", startOfYear, endOfYear).
		Where("time_events.deleted_at IS NULL").
		Where("employees.deleted_at IS NULL")

	// Применяем фильтры
	if searchParams.Email != "" {
		query = query.Where("employees.email = ?", searchParams.Email)
	}
	if searchParams.DepartmentID != 0 {
		query = query.Where("employees.department_id = ?", searchParams.DepartmentID)
	}
	if searchParams.SearchQuery != "" {
		searchPattern := "%" + searchParams.SearchQuery + "%"
		query = query.Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ?",
			searchPattern, searchPattern)
	}

	// Сортировка и пагинация
	query = query.Order("time_events.updated_at DESC").
		Offset(searchParams.Offset).
		Limit(searchParams.Limit)

	// Выполняем запрос
	err = query.Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	return events, totalCount, nil
}

func (repo *UserRepository) GetLastAddEvents(searchParams SearchParam) ([]StatusPeriod, int64, error) {
	var history []StatusPeriod

	currentYear := time.Now().Year()
	startOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(currentYear, 12, 31, 23, 59, 59, 0, time.UTC)

	// Получаем общее количество
	totalCount, err := repo.getStatusPeriodsCount(searchParams, startOfYear, endOfYear)
	if err != nil {
		return nil, 0, err
	}

	// Основной запрос с Preload
	query := repo.DataBase.DB.
		Preload("Employee", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL")
		}).
		Preload("WhoAdded", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL")
		}).
		Preload("StatusType", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL")
		}).
		Joins("INNER JOIN employees ON status_periods.employee_id = employees.id").
		Where("status_periods.start_date BETWEEN ? AND ?", startOfYear, endOfYear).
		Where("status_periods.deleted_at IS NULL").
		Where("employees.deleted_at IS NULL")

	// Применяем фильтры
	if searchParams.Email != "" {
		query = query.Where("employees.email = ?", searchParams.Email)
	}
	if searchParams.DepartmentID != 0 {
		query = query.Where("employees.department_id = ?", searchParams.DepartmentID)
	}
	if searchParams.SearchQuery != "" {
		searchPattern := "%" + searchParams.SearchQuery + "%"
		query = query.Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ?",
			searchPattern, searchPattern)
	}

	// Сортировка и пагинация
	query = query.Order("status_periods.updated_at DESC")

	query = query.Offset(searchParams.Offset).Limit(searchParams.Limit)

	// Выполняем запрос
	err = query.Find(&history).Error
	if err != nil {
		return nil, 0, err
	}

	return history, totalCount, nil
}

//

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

func (repo *UserRepository) GetLastTimeEvents(email string, limit ...int) ([]TimeEvent, error) {
	var events []TimeEvent

	// Текущий год
	currentYear := time.Now().Year()
	startOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(currentYear, 12, 31, 23, 59, 59, 0, time.UTC)

	query := repo.DataBase.DB.
		Preload("Employee").
		Preload("WhoAdded").
		Preload("EventType").
		Joins("INNER JOIN employees ON time_events.employee_id = employees.id").
		Where("employees.email = ?", email).
		Where("time_events.date BETWEEN ? AND ?", startOfYear, endOfYear).
		Order("time_events.updated_at DESC")

	if len(limit) > 0 && limit[0] > 0 {
		query = query.Limit(limit[0])
	}

	err := query.Find(&events).Error
	return events, err
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

func (repo *UserRepository) GetTimeEventHistory(
	email string,
	timeStart, timeEnd time.Time,
) ([]TimeEvent, error) {
	var history []TimeEvent

	err := repo.DataBase.DB.
		Preload("Employee").
		Preload("WhoAdded").
		Preload("EventType"). // Загружаем связанный EventType
		Joins("INNER JOIN employees ON time_events.employee_id = employees.id").
		Where("employees.email = ?", email).
		Where("time_events.date >= ?", timeStart).
		Where("time_events.date <= ?", timeEnd).
		Order("time_events.date DESC").
		Find(&history).
		Error

	return history, err
}

type UserUpdateData struct {
	FirstName      string
	LastName       string
	Position       string
	Department     string
	IsActive       bool
	IsAdmin        bool
	ShowTimeEvents bool
}

func (repo *UserRepository) UpdateUserProfile(email string, data UserUpdateData) error {
	var departmentID uint

	if data.Department != "" {
		department, err := repo.FindOrCreateDepartment(data.Department)
		if err != nil {
			return err
		}
		departmentID = department.ID
	}

	return repo.DataBase.Model(&Employee{}).
		Where("email = ?", email).
		Updates(map[string]any{
			"first_name":       data.FirstName,
			"last_name":        data.LastName,
			"position":         data.Position,
			"department_id":    departmentID,
			"is_active":        data.IsActive,
			"is_admin":         data.IsAdmin,
			"updated_at":       time.Now(),
			"show_time_events": data.ShowTimeEvents,
		}).Error
}

func (repo *UserRepository) GetCountUsersByDepartment(departmentId int) (int64, error) {
	var count int64
	var err error

	if departmentId == 0 {
		// Считаем всех пользователей
		err = repo.DataBase.Model(&Employee{}).Count(&count).Error
	} else {
		// Считаем пользователей конкретного отдела
		err = repo.DataBase.Model(&Employee{}).
			Where("department_id = ?", departmentId).
			Count(&count).Error
	}

	return count, err
}

type SearchParam struct {
	Email        string
	DepartmentID uint
	SearchQuery  string
	Offset       int
	Limit        int
}

func (repo *UserRepository) GetUsersByParam(searchParam SearchParam) ([]Employee, int64, error) {
	var employees []Employee
	var totalCount int64

	// Создаем базовый запрос
	query := repo.DataBase.Model(&Employee{}).Preload("Department")

	// Фильтр по departmentID (если не 0)
	if searchParam.DepartmentID != 0 {
		query = query.Where("department_id = ?", searchParam.DepartmentID)
	}

	// Поиск по ФИО (если SearchQuery не пустой)
	if searchParam.SearchQuery != "" {
		searchPattern := "%" + searchParam.SearchQuery + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Сортировка по алфавитному порядку (фамилия, имя)
	query = query.Order("is_active DESC, last_name ASC, first_name ASC")

	// Считаем общее количество записей (для пагинации)
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Применяем пагинацию
	query = query.Offset(searchParam.Offset).Limit(searchParam.Limit)

	// Выполняем запрос
	if err := query.Find(&employees).Error; err != nil {
		return nil, 0, err
	}

	return employees, totalCount, nil
}

type TimeEventStat struct {
	LatelyMin       int
	LatelyCount     int
	EarlyLeaveMin   int
	EarlyLeaveCount int
}

func (repo *UserRepository) GetYearTimeEventStat(email string) (TimeEventStat, error) {
	var stat TimeEventStat

	// Текущий год
	year := time.Now().Year()
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	// Находим сотрудника
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", email).First(&employee).Error; err != nil {
		return stat, fmt.Errorf("сотрудник не найден: %w", err)
	}

	// Загружаем все временные события за год для этого сотрудника
	var events []TimeEvent
	err := repo.DataBase.DB.
		Preload("EventType").
		Where("employee_id = ? AND date BETWEEN ? AND ?", employee.ID, startOfYear, endOfYear).
		Find(&events).Error
	if err != nil {
		return stat, fmt.Errorf("ошибка загрузки событий за год: %w", err)
	}

	// Подсчитываем в Go
	for _, event := range events {
		switch event.EventType.Code {
		case "late":
			stat.LatelyCount++
			if event.DifferenceMin > 0 {
				stat.LatelyMin += event.DifferenceMin
			}
		case "early_leave":
			stat.EarlyLeaveCount++
			if event.DifferenceMin > 0 {
				stat.EarlyLeaveMin += event.DifferenceMin // делаем положительным
			}
		}
	}

	return stat, nil
}

func (repo *UserRepository) GetTimeEventStat(month int, email string) (TimeEventStat, error) {
	var stat TimeEventStat

	// Текущий год
	year := time.Now().Year()

	// Начало и конец месяца
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second) // последняя секунда месяца

	// Находим ID сотрудника
	var employee Employee
	if err := repo.DataBase.DB.Where("email = ?", email).First(&employee).Error; err != nil {
		return stat, fmt.Errorf("сотрудник не найден: %w", err)
	}

	// Загружаем ВСЕ временные события за месяц для этого сотрудника
	var events []TimeEvent
	err := repo.DataBase.DB.
		Preload("EventType").
		Where("employee_id = ? AND date BETWEEN ? AND ?", employee.ID, startOfMonth, endOfMonth).
		Find(&events).Error
	if err != nil {
		return stat, fmt.Errorf("ошибка загрузки событий: %w", err)
	}

	// Подсчитываем в Go
	for _, event := range events {
		switch event.EventType.Code {
		case "late":
			stat.LatelyCount++
			if event.DifferenceMin > 0 {
				stat.LatelyMin += event.DifferenceMin
			}
		case "early_leave":
			stat.EarlyLeaveCount++
			if event.DifferenceMin > 0 {
				stat.EarlyLeaveMin += event.DifferenceMin
			}
		}
	}

	return stat, nil
}

func (repo *UserRepository) IsAdmin(email string) bool {
	var user Employee

	// Выполняем запрос к базе данных
	result := repo.DataBase.Model(&Employee{}).
		Select("is_admin").
		Where("email = ?", email).
		First(&user)

	// Если произошла ошибка или пользователь не найден, возвращаем false
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Пользователь не найден
			return false
		}
		// Другая ошибка базы данных
		// Можно залогировать ошибку: log.Printf("Database error: %v", result.Error)
		return false
	}

	return user.IsAdmin
}

func (repo *UserRepository) GetShowTimeEvents(email string) bool {
	var user Employee

	// Выполняем запрос к базе данных
	result := repo.DataBase.Model(&Employee{}).
		Select("show_time_events").
		Where("email = ?", email).
		First(&user)

	// Если произошла ошибка или пользователь не найден, возвращаем false
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Пользователь не найден
			return false
		}
		// Другая ошибка базы данных
		// Можно залогировать ошибку: log.Printf("Database error: %v", result.Error)
		return false
	}

	return user.ShowTimeEvents
}
