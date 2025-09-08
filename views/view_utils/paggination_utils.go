package viewutils

import (
	"net/url"
	"strconv"
)

// Функция для формирования URL с сохранением параметров
func BuildPageURL(page int, queryParams map[string]string) string {
	params := url.Values{}

	// Копируем все существующие параметры
	for key, value := range queryParams {
		if key != "page" && value != "" {
			params.Add(key, value)
		}
	}

	// Добавляем новый номер страницы
	params.Set("page", strconv.Itoa(page))

	return "/users_activity?" + params.Encode()
}
