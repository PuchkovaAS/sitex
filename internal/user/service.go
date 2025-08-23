package user

import (
	"time"
)

type UserServiceDeps struct {
	UserRepository UserRepository
}

type UserService struct {
	userRepository UserRepository
}

func NewUserService(deps *UserServiceDeps) *UserService {
	return &UserService{userRepository: deps.UserRepository}
}

func (service *UserService) getHistory(
	email string,
	timeFrom, timeTo time.Time,
) ([]StatusHistoryResponse, error) {
	history, err := service.userRepository.GetStatusHistory(email, timeFrom, timeTo)
	if err != nil {
		return nil, err
	}

	// Преобразуем в DTO
	var response []StatusHistoryResponse
	for _, period := range history {
		response = append(response, period.ToStatusHistoryResponse())
	}
	return response, nil
}

type DayStatus struct {
	Date    int    // День месяца
	Status  string // Статус
	Comment string // Комментарий
}

type StatusCount struct {
	Status string
	Count  int
}

// Метод для расчета дней
func (service *UserService) calcDays(
	timeFrom, timeTo time.Time,
	historyStatus []StatusHistoryResponse,
) ([]DayStatus, []StatusCount) {
	// Используем полную дату как ключ (год-месяц-день)
	statusMap := make(map[string]DayStatus)

	// Заполняем мапу данными из истории
	for _, status := range historyStatus {
		key := status.StartDate.Format("2006-01-02")
		statusMap[key] = DayStatus{
			Date:    status.StartDate.Day(),
			Status:  status.StatusName,
			Comment: status.Comment,
		}
	}

	// Создаем слайс для хранения всех дней периода
	var result []DayStatus
	currentDate := timeFrom
	lastStatus := "В офисе"

	// Проходим по всем дням в диапазоне
	for currentDate.Before(timeTo) || currentDate.Equal(timeTo) {
		key := currentDate.Format("2006-01-02")

		// Если день есть в мапе - берем его статус
		if status, ok := statusMap[key]; ok {
			result = append(result, status)
			lastStatus = status.Status
		} else {
			// Если статуса нет - определяем статус по умолчанию
			statusName := lastStatus

			// Проверяем выходные
			if currentDate.Weekday() == time.Saturday || currentDate.Weekday() == time.Sunday {
				statusName = "Выходной"
			}

			result = append(result, DayStatus{
				Date:    currentDate.Day(),
				Status:  statusName,
				Comment: "",
			})
		}

		currentDate = currentDate.AddDate(0, 0, 1) // Следующий день
	}

	// Подсчитываем количество дней по статусам
	statusCounter := make(map[string]int)
	for _, day := range result {
		statusCounter[day.Status]++
	}

	// Преобразуем счетчик в слайс
	var statusCounts []StatusCount
	for status, count := range statusCounter {
		statusCounts = append(statusCounts, StatusCount{
			Status: status,
			Count:  count,
		})
	}

	return result, statusCounts
}

func (service *UserService) GetDaysStatus(
	email string,
	timeTo, timeFrom time.Time,
) ([]DayStatus, []StatusCount, error) {
	history, err := service.getHistory(email, timeTo, timeFrom)
	if err != nil {
		return nil, nil, err
	}
	dayStatus, statusCount := service.calcDays(timeTo, timeFrom, history)
	return dayStatus, statusCount, nil
}

func (service *UserService) GetDateRange() (time.Time, time.Time) {
	now := time.Now()

	// Первый день текущего месяца
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Последний день текущего месяца
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, now.Location())

	return firstDay, lastDay
}
