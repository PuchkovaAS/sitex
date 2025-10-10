package viewutils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/a-h/templ"
)

func GetMonthName(monthNum int) string {
	months := []string{
		"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
		"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
	}
	return months[monthNum-1]
}

func GetStatusClass(status string) string {
	switch status {
	// ──────────────── ОТДЫХ / ОТПУСК ────────────────
	case "Отпуск":
		return "text-blue-700 bg-blue-100 border border-blue-200"
	case "Отпуск за свой счёт":
		return "text-violet-700 bg-violet-100 border border-violet-200"
	case "Выходной":
		return "text-gray-500 bg-gray-100 border border-gray-200"

	// ──────────────── РАБОТА ────────────────
	case "В офисе":
		return "text-emerald-700 bg-emerald-100 border border-emerald-200"
	case "Удаленная работа":
		return "text-orange-600 bg-orange-100 border border-orange-200"
	case "Работа в выходной день":
		return "text-indigo-600 bg-indigo-100 border border-indigo-200"

	// ──────────────── ОТСУТСТВИЕ (не по желанию) ────────────────
	case "Больничный":
		return "text-rose-700 bg-rose-100 border border-rose-200"

	// ──────────────── ДОПОЛНИТЕЛЬНЫЕ ДНИ / КОМПЕНСАЦИИ ────────────────
	case "Отгул":
		return "text-amber-700 bg-amber-100 border border-amber-200"
	case "Командировка":
		return "text-coral-600 bg-coral-100 border border-coral-200"

	// ──────────────── ПО УМОЛЧАНИЮ ────────────────
	default:
		return "text-gray-400 bg-white border border-gray-200"
	}
}

// Вспомогательная функция для определения активного класса
func IsActive(currentPath, targetPath string) string {
	if currentPath == targetPath {
		return "text-gray-900 bg-gray-100"
	}
	return "text-gray-700 hover:bg-gray-100"
}

// Для ссылок, которые могут иметь подпути
func IsActivePrefix(currentPath, targetPath string) string {
	if strings.HasPrefix(currentPath, targetPath) {
		return "text-gray-900 bg-gray-100"
	}
	return "text-gray-700 hover:bg-gray-100"
}

// Вспомогательные функции (можно вынести в view_utils если нужно)
// GetDisplayName возвращает краткое или адаптированное отображаемое имя статуса.
func GetDisplayName(status string) string {
	switch status {
	case "Удаленная работа":
		return "Удалённо"
	case "Работа в выходной день":
		return "Выходной\nработа"
	case "Отпуск за свой счёт":
		return "За свой счёт"
	default:
		return status
	}
}

// Используется, например, в иконках или компактных метках.
func GetTextColorClass(status string) string {
	switch status {
	// Отдых
	case "Отпуск":
		return "text-blue-700"
	case "Отпуск за свой счёт":
		return "text-violet-700"
	case "Выходной":
		return "text-gray-500"

	// Работа
	case "В офисе":
		return "text-emerald-700"
	case "Удаленная работа":
		return "text-orange-600"
	case "Работа в выходной день":
		return "text-indigo-600"

	// Отсутствие
	case "Больничный":
		return "text-rose-700"

	// Дополнительно
	case "Отгул":
		return "text-amber-700"
	case "Командировка":
		return "text-coral-600"

	default:
		return "text-gray-400"
	}
}

func GetInitals(firstName, lastName string) string {
	// Безопасное получение инициалов
	initials := ""

	// Получаем первую руну имени
	if firstName != "" {
		runes := []rune(firstName)
		if len(runes) > 0 {
			initials += string(runes[0])
		}
	}

	// Получаем первую руну фамилии
	if lastName != "" {
		runes := []rune(lastName)
		if len(runes) > 0 {
			initials += string(runes[0])
		}
	}

	// Если оба поля пустые
	if initials == "" {
		initials = "U" // User по умолчанию
	}

	initials = strings.ToUpper(initials)

	return initials
}

// Вспомогательные функции (должны быть реализованы в Go коде)
func buildPageURL(page, deptID int, searchQuery string) string {
	// Реализация построения URL с параметрами
	return fmt.Sprintf("/activity?page=%d&department=%d&search=%s", page, deptID, url.QueryEscape(searchQuery))
}

func generatePaginationLinks() templ.Component {
	// Реализация генерации ссылок пагинации
	return templ.Raw("") // Заглушка
}
