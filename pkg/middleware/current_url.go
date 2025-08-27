package middleware

import "github.com/gofiber/fiber/v2"

func CurrentURLMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Получаем текущий путь
		currentPath := c.Path()

		// Сохраняем в контекст
		c.Locals("currentPath", currentPath)

		// Передаем управление следующему обработчику
		return c.Next()
	}
}
