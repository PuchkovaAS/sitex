package user

import "time"

// StatusHistoryResponse DTO для ответа с историей статусов
type StatusHistoryResponse struct {
	ID           uint      `json:"id"`
	StartDate    time.Time `json:"start_date"`
	Comment      string    `json:"comment"`
	StatusName   string    `json:"status_name"`
	OneTimeEvent bool      `json:"one_time_event"`
}

// ToStatusHistoryResponse преобразует StatusPeriod в DTO
func (sp *StatusPeriod) ToStatusHistoryResponse() StatusHistoryResponse {
	return StatusHistoryResponse{
		ID:           sp.ID,
		StartDate:    sp.StartDate,
		Comment:      sp.Comment,
		StatusName:   sp.StatusType.Name, // Берем название из связанной таблицы
		OneTimeEvent: sp.OneTimeEvent,
	}
}

type TimeEventHistoryResponse struct {
	ID          uint      `json:"id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	EventName   string    `json:"event_name"`
}

// ToStatusHistoryResponse преобразует StatusPeriod в DTO
func (te *TimeEvent) ToTimeEventHistoryResponse() TimeEventHistoryResponse {
	return TimeEventHistoryResponse{
		ID:          te.ID,
		Date:        te.Date,
		Description: te.Description,
		EventName:   te.EventType.Name, // Берем название из связанной таблицы
	}
}
