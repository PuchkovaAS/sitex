package user

import (
	"fmt"
	"sitex/pkg/calendar"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type UserServiceDeps struct {
	UserRepository UserRepository
	CustomLogger   *zerolog.Logger
	WorkCalendar   *calendar.HolidayCalendar
}

type UserService struct {
	userRepository UserRepository
	customLogger   *zerolog.Logger
	workCalendar   *calendar.HolidayCalendar
}

func NewUserService(deps *UserServiceDeps) *UserService {
	return &UserService{
		userRepository: deps.UserRepository,
		customLogger:   deps.CustomLogger,
		workCalendar:   deps.WorkCalendar,
	}
}

func (service *UserService) getTimeEventHistory(
	email string,
	timeStart, timeEnd time.Time,
) ([]TimeEventHistoryResponse, error) {
	history, err := service.userRepository.GetTimeEventHistory(email, timeStart, timeEnd)
	if err != nil {
		return nil, err
	}
	formatTime := func(t string) string {
		if len(t) >= 5 {
			return t[:5] // "09:15:30" → "09:15"
		}
		return t
	}

	// Мапа: дата → агрегированная запись
	eventMap := make(map[string]*TimeEventHistoryResponse)

	for _, event := range history {
		key := event.Date.Format("2006-01-02")

		if existing, ok := eventMap[key]; ok {
			// Уже есть запись на эту дату — объединяем
			if event.Description != "" {
				if existing.Description != "" {
					existing.Description += "\n" + event.Description
				} else {
					existing.Description = event.Description
				}
			}

			desc := fmt.Sprintf("%s, План: %s, Факт: %s",
				event.EventType.Name,
				formatTime(event.ScheduledTime),
				formatTime(event.ActualTime),
			)
			// Опционально: объединить EventName
			if event.EventType.Name != "" && !strings.Contains(existing.EventName, event.EventType.Name) {
				if existing.EventName != "" {
					existing.EventName += "\n" + desc
				} else {
					existing.EventName = desc
				}
			}
		} else {
			// Первая запись на эту дату
			resp := event.ToTimeEventHistoryResponse()
			eventMap[key] = &resp
			desc := fmt.Sprintf("%s, План: %s, Факт: %s",
				event.EventType.Name,
				formatTime(event.ScheduledTime),
				formatTime(event.ActualTime),
			)
			eventMap[key].EventName = desc
		}
	}

	// Преобразуем мапу в слайс и сортируем по дате (опционально)
	var response []TimeEventHistoryResponse
	for _, event := range eventMap {
		response = append(response, *event)
	}

	// Сортировка по дате (от новых к старым)
	sort.Slice(response, func(i, j int) bool {
		return response[i].Date.After(response[j].Date)
	})

	return response, nil
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
	Date         int
	Status       string
	Comment      string
	OneTimeEvent bool
	IsAddStatus  bool
	IsTimeEvent  bool
}

type CalcDaysDeps struct {
	Email            string
	TimeStart        time.Time
	TimeEnd          time.Time
	HistoryStatus    []StatusHistoryResponse
	TimeEventHistory []TimeEventHistoryResponse
}

func (service *UserService) calcDays(deps CalcDaysDeps) ([]DayStatus, map[string]int) {
	// Используем полную дату как ключ (год-месяц-день)
	statusMap := make(map[string]DayStatus)
	timeEventMap := make(map[string]DayStatus)

	// Заполняем мапу данными из истории
	for _, status := range deps.HistoryStatus {
		key := status.StartDate.Format("2006-01-02")
		statusMap[key] = DayStatus{
			Date:         status.StartDate.Day(),
			Status:       status.StatusName,
			Comment:      status.Comment,
			OneTimeEvent: status.OneTimeEvent,
			IsAddStatus:  true,
		}
	}

	// Заполняем мапу данными из истории
	for _, status := range deps.TimeEventHistory {
		key := status.Date.Format("2006-01-02")
		timeEventMap[key] = DayStatus{
			Comment:     status.EventName + "\t" + status.Description,
			IsTimeEvent: true,
		}
	}

	// Создаем слайс для хранения всех дней периода
	var result []DayStatus
	currentDate := deps.TimeStart

	var lastStatus string
	lastStatus, err := service.userRepository.GetLastStatus(
		deps.Email,
		deps.TimeStart.Add(-24*time.Hour).Truncate(24*time.Hour),
	)
	if err != nil {
		lastStatus = "В офисе"
	}

	lastComment := ""

	today := time.Now()
	// Проходим по всем дням в диапазоне
	for currentDate.Before(deps.TimeEnd) || currentDate.Equal(deps.TimeEnd) {
		if currentDate.After(today) {
			result = append(result, DayStatus{
				Date:    currentDate.Day(),
				Status:  "",
				Comment: "",
			})

			currentDate = currentDate.AddDate(0, 0, 1) // Следующий день
			continue
		}

		key := currentDate.Format("2006-01-02")

		isAddStatus := false

		// Если день есть в мапе - берем его статус
		if status, ok := statusMap[key]; ok {

			if !status.OneTimeEvent {

				lastStatus = status.Status
				lastComment = status.Comment
			}

			if timeEvent, ok := timeEventMap[key]; ok {
				status.IsTimeEvent = true
				status.Comment = status.Comment + "\n" + timeEvent.Comment
			}
			result = append(result, status)
		} else {
			// Если статуса нет - определяем статус по умолчанию
			statusName := lastStatus
			statusComment := lastComment

			isWeekday, comment := service.workCalendar.IsWorkingDay(currentDate)

			if !isWeekday {
				statusName = "Выходной"
				statusComment = comment
			} else if statusComment != "" {
				statusComment = statusComment + "\n" + comment
			}

			if timeEvent, ok := timeEventMap[key]; ok {
				result = append(result, DayStatus{
					Date:        currentDate.Day(),
					Status:      statusName,
					Comment:     statusComment + "\n" + timeEvent.Comment,
					IsTimeEvent: true,
					IsAddStatus: isAddStatus,
				})
			} else {
				result = append(result, DayStatus{
					Date:        currentDate.Day(),
					Status:      statusName,
					Comment:     statusComment,
					IsAddStatus: isAddStatus,
				})
			}

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
	ShowTimeEvents bool,
) ([]DayStatus, map[string]int, error) {
	history, err := service.getHistory(email, timeStart, timeEnd)
	if err != nil {
		return nil, nil, err
	}
	var timeEvents []TimeEventHistoryResponse
	if ShowTimeEvents {

		timeEvents, err = service.getTimeEventHistory(email, timeStart, timeEnd)
		if err != nil {
			return nil, nil, err
		}
	}
	dayStatus, statusCount := service.calcDays(CalcDaysDeps{
		Email:            email,
		TimeStart:        timeStart,
		TimeEnd:          timeEnd,
		HistoryStatus:    history,
		TimeEventHistory: timeEvents,
	})
	return dayStatus, statusCount, nil
}

func (service *UserService) GetDateRange(now time.Time) (time.Time, time.Time) {
	// Первый день текущего месяца
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Последний день текущего месяца
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, now.Location())

	return firstDay, lastDay
}

func (service *UserService) GetYearHistory(
	email string,
	ShowTimeEvents bool,
) ([]MonthHistory, map[string]int, error) {
	var resultHistory []MonthHistory
	statusCount := make(map[string]int)

	// Получаем текущий год
	currentYear := time.Now().Year()

	for month := 1; month <= 12; month++ {
		startOfMonth := time.Date(currentYear, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

		timeStart, timeEnd := service.GetDateRange(startOfMonth)

		daysStatus, monthStatusCount, err := service.GetDaysStatus(email, timeStart, timeEnd, ShowTimeEvents)
		if err != nil {
			return nil, nil, err
		}

		// Накопление статистики за год (суммируем значения)
		for status, count := range monthStatusCount {
			statusCount[status] += count
		}

		// Получаем русское название месяца
		monthName := getRussianMonthName(startOfMonth.Month())
		weekday := startOfMonth.Weekday()
		offset := int(weekday)
		if offset == 0 { // воскресенье
			offset = 6
		} else {
			offset = offset - 1
		}

		resultHistory = append(resultHistory, MonthHistory{
			Name:              monthName,
			Number:            int(startOfMonth.Month()),
			WeekdayFirstMonth: offset,
			HistoryStatus:     daysStatus,
		})
	}
	return resultHistory, statusCount, nil
}

func (service *UserService) GetMonthHistory(
	month int,
	email string,
	countMonth int,
	ShowTimeEvents bool,
) ([]MonthHistory, map[string]int, error) {
	var resultHistory []MonthHistory
	statusCount := make(map[string]int)

	if month < countMonth {
		countMonth = month
	}

	currentTime := time.Date(time.Now().Year(), time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	for i := range countMonth {
		// Вычисляем начало месяца (смещение на i месяцев назад)
		targetDate := currentTime.AddDate(0, -countMonth+i+1, 0)
		startOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)

		timeStart, timeEnd := service.GetDateRange(startOfMonth)

		daysStatus, monthStatusCount, err := service.GetDaysStatus(email, timeStart, timeEnd, ShowTimeEvents)
		if err != nil {
			return nil, nil, err
		}

		statusCount = monthStatusCount

		// Получаем русское название месяца
		monthName := getRussianMonthName(startOfMonth.Month())
		weekday := startOfMonth.Weekday()
		offset := int(weekday)
		if offset == 0 { // воскресенье
			offset = 6
		} else {
			offset = offset - 1
		}

		resultHistory = append(resultHistory, MonthHistory{
			Name:              monthName,
			Number:            int(startOfMonth.Month()),
			WeekdayFirstMonth: offset,
			HistoryStatus:     daysStatus,
		})
	}

	return resultHistory, statusCount, nil
}

func getRussianMonthName(month time.Month) string {
	months := map[time.Month]string{
		time.January:   "Январь",
		time.February:  "Февраль",
		time.March:     "Март",
		time.April:     "Апрель",
		time.May:       "Май",
		time.June:      "Июнь",
		time.July:      "Июль",
		time.August:    "Август",
		time.September: "Сентябрь",
		time.October:   "Октябрь",
		time.November:  "Ноябрь",
		time.December:  "Декабрь",
	}
	return months[month]
}
