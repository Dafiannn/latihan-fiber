package middlewares

import (
	"github.com/dafian/siakad-mini/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// IsAdmin adalah middleware untuk mengecek apakah user memiliki role admin
func IsAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "admin" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak. Fitur ini hanya untuk Admin.", nil)
		}
		return c.Next()
	}
}
