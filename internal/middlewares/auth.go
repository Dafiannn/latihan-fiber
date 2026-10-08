package middlewares

import (
	"fmt"
	"os"
	"strings"

	"github.com/dafian/siakad-mini/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected adalah middleware yang mewajibkan token JWT valid (Error 401)
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Token tidak ada atau salah", nil)
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode unauth")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Token salah atau kedaluwarsa", nil)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid token claims", nil)
		}

		// Simpan data token ke variabel Locals Fiber agar bisa dipakai di Handler
		c.Locals("user_id", uint(claims["user_id"].(float64)))
		c.Locals("role", claims["role"].(string))

		return c.Next()
	}
}
