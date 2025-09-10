package middleware

import (
	"sitex/pkg/intergaces"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func IsAdminMiddleware(store *session.Store, repo intergaces.IUserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Redirect("/login")
		}

		email, ok := sess.Get("email").(string)
		if !ok || email == "" {
			return c.Redirect("/login")
		}

		isAdmin := repo.IsAdmin(email)

		c.Locals("email", email)

		if !isAdmin {
			return c.Redirect("/errors/403") // перенаправление на страницу
		}

		return c.Next()
	}
}
