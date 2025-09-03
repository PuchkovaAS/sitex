package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func IsAdminMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Redirect("/login")
		}

		email, ok := sess.Get("email").(string)
		if !ok || email == "" {
			return c.Redirect("/login")
		}

		// Проверяем, есть ли уже информация об админских правах в сессии
		isAdmin, exists := sess.Get("is_admin").(bool)

		// Если нет в сессии - запрашиваем из БД и сохраняем
		if !exists || !isAdmin {
			return c.Redirect("/errors/403") // перенаправление на страницу
		}

		c.Locals("email", email)

		return c.Next()
	}
}
