package viewutils

func GetStatusClass(status string) string {
	switch status {
	// Отдых - серые оттенки
	case "В отпуске":
		return "text-purple-700 bg-purple-100 border border-purple-300"

	case "Выходной":
		return "text-gray-500 bg-gray-200 border border-gray-400"

	// Работа - зеленые оттенки
	case "В офисе":
		return "text-green-700 bg-green-100 border border-green-300"

	// Больничный - красные оттенки
	case "Больничный":
		return "text-red-700 bg-red-100 border border-red-300"

	// Отгул - желтые оттенки
	case "Отгул":
		return "text-yellow-700 bg-yellow-100 border border-yellow-300"

	// Удаленная работа - оранжевые оттенки
	case "Удаленная работа":
		return "text-orange-700 bg-orange-100 border border-orange-300"

	// Работа в выходной - фиолетовые оттенки
	case "Работа в выходной день":
		return "text-purple-700 bg-purple-100 border border-purple-300"
	// Командировка - розовые оттенки
	case "Командировка":
		return "text-pink-700 bg-pink-100 border border-pink-300"

	default:
		return "text-gray-400 bg-gray-100 border border-gray-300"
	}
}
