package viewutils

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
