package user

import "time"

// StatusHistoryResponse DTO для ответа с историей статусов
type StatusHistoryResponse struct {
	ID         uint      `json:"id"`
	StartDate  time.Time `json:"start_date"`
	Comment    string    `json:"comment"`
	StatusName string    `json:"status_name"`
}

// ToStatusHistoryResponse преобразует StatusPeriod в DTO
func (sp *StatusPeriod) ToStatusHistoryResponse() StatusHistoryResponse {
	return StatusHistoryResponse{
		ID:         sp.ID,
		StartDate:  sp.StartDate,
		Comment:    sp.Comment,
		StatusName: sp.StatusType.Name, // Берем название из связанной таблицы
	}
}
