package routes

import (
	"github.com/dafian/siakad-mini/internal/handlers"
	"github.com/dafian/siakad-mini/internal/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Prefix /api/v1
	api := app.Group("/api/v1")

	// 1 & 2. Auth Routes
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Get("/me", middlewares.Protected(), handlers.Me)
}
