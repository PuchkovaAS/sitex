package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func AuthMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Redirect("/login")
		}

		email, ok := sess.Get("email").(string)
		if !ok || email == "" {
			if err := sess.Save(); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			return c.Redirect("/login")
		}

		c.Locals("email", email)
		return c.Next()
	}
}
