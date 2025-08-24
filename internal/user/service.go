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
	timeStart, timeEnd time.Time,
) ([]StatusHistoryResponse, error) {
	history, err := service.userRepository.GetStatusHistory(email, timeStart, timeEnd)
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

type CalcDaysDeps struct {
	Email         string
	TimeStart     time.Time
	TimeEnd       time.Time
	HistoryStatus []StatusHistoryResponse
}

func (service *UserService) calcDays(deps CalcDaysDeps) ([]DayStatus, map[string]int) {
	// Используем полную дату как ключ (год-месяц-день)
	statusMap := make(map[string]DayStatus)

	// Заполняем мапу данными из истории
	for _, status := range deps.HistoryStatus {
		key := status.StartDate.Format("2006-01-02")
		statusMap[key] = DayStatus{
			Date:    status.StartDate.Day(),
			Status:  status.StatusName,
			Comment: status.Comment,
		}
	}

	// Создаем слайс для хранения всех дней периода
	var result []DayStatus
	currentDate := deps.TimeStart

	var lastStatus string
	lastStatus, err := service.userRepository.GetCurrentStatus(
		deps.Email,
		deps.TimeStart.Add(-24*time.Hour).Truncate(24*time.Hour),
	)
	if err != nil {
		lastStatus = "В офисе"
	}

	// Проходим по всем дням в диапазоне
	for currentDate.Before(deps.TimeEnd) || currentDate.Equal(deps.TimeEnd) {
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

	return result, statusCounter
}

func (service *UserService) GetDaysStatus(
	email string,
	timeStart, timeEnd time.Time,
) ([]DayStatus, map[string]int, error) {
	history, err := service.getHistory(email, timeStart, timeEnd)
	if err != nil {
		return nil, nil, err
	}
	dayStatus, statusCount := service.calcDays(CalcDaysDeps{
		Email:         email,
		TimeStart:     timeStart,
		TimeEnd:       timeEnd,
		HistoryStatus: history,
	})
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
