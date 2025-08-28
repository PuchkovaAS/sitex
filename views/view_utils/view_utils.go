package viewutils

import "strings"

func GetMonthName(monthNum int) string {
	months := []string{
		"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
		"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
	}
	return months[monthNum-1]
}

func GetStatusClass(status string) string {
	switch status {
	// Отдых - спокойные синие
	case "Отпуск":
		return "text-blue-700 bg-blue-100 border border-blue-200"

	case "Выходной":
		return "text-gray-500 bg-gray-100 border border-gray-200"

	// Работа - естественные зеленые
	case "В офисе":
		return "text-emerald-700 bg-emerald-100 border border-emerald-200"

	// Больничный - мягкие красные
	case "Больничный":
		return "text-rose-700 bg-rose-100 border border-rose-200"

	// Отгул - теплые янтарные
	case "Отгул":
		return "text-amber-700 bg-amber-100 border border-amber-200"

	// Удаленная работа - терракотовые
	case "Удаленная работа":
		return "text-orange-600 bg-orange-100 border border-orange-200"

	// Работа в выходной - лавандовые
	case "Работа в выходной день":
		return "text-indigo-600 bg-indigo-100 border border-indigo-200"

	// Командировка - коралловые
	case "Командировка":
		return "text-coral-600 bg-coral-100 border border-coral-200"

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
func GetDisplayName(status string) string {
	switch status {
	case "Удаленная работа":
		return "Удалённо"
	case "Работа в выходной день":
		return "Выходной\nработа"
	default:
		return status
	}
}

func GetTextColorClass(status string) string {
	switch status {
	case "Отпуск":
		return "text-blue-700"
	case "Выходной":
		return "text-gray-500"
	case "В офисе":
		return "text-emerald-700"
	case "Больничный":
		return "text-rose-700"
	case "Отгул":
		return "text-amber-700"
	case "Удаленная работа":
		return "text-orange-600"
	case "Работа в выходной день":
		return "text-indigo-600"
	case "Командировка":
		return "text-coral-600"
	default:
		return "text-gray-400"
	}
}
