package templeadapter

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

func Render(c *fiber.Ctx, component templ.Component, code int) error {
	c.Response().Header.SetContentType("text/html")
	c.Status(code)
	return component.Render(c.Context(), c.Response().BodyWriter())
}
