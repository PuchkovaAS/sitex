package calendar

import (
	"encoding/json"
	"os"
	"time"
)

type Holiday struct {
	Date        string `json:"date"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Transfer struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Description string `json:"description"`
}

type WorkingWeekend struct {
	Date        string `json:"date"`
	Description string `json:"description"`
}

type HolidayCalendar struct {
	Year            int              `json:"year"`
	Country         string           `json:"country"`
	Holidays        []Holiday        `json:"holidays"`
	Transfers       []Transfer       `json:"transfers"`
	WorkingWeekends []WorkingWeekend `json:"working_weekends"`
}

func LoadHolidays(filename string) (*HolidayCalendar, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var calendar HolidayCalendar
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&calendar)
	if err != nil {
		return nil, err
	}

	return &calendar, nil
}

func (workCalendar *HolidayCalendar) IsWorkingDay(
	date time.Time,
) (bool, string) {
	dateStr := date.Format("2006-01-02")

	// 1. Проверяем, не является ли день перенесенным выходным
	for _, transfer := range workCalendar.Transfers {
		if transfer.To == dateStr {
			return false, "Перенесенный выходной день: " + transfer.Description
		}
	}

	// 2. Проверяем, не является ли день перенесенным рабочим днем
	for _, transfer := range workCalendar.Transfers {
		if transfer.From == dateStr {
			return true, "Перенесенный рабочий день: " + transfer.Description
		}
	}

	// 3. Проверяем, является ли день стандартным выходным (сб/вс)
	weekday := date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false, "Выходной день (суббота/воскресенье)"
	}

	// 4. Проверяем, не является ли день праздником
	for _, holiday := range workCalendar.Holidays {
		if holiday.Date == dateStr {
			return false, "Праздничный день: " + holiday.Name
		}
	}

	return true, "Рабочий день"
}
