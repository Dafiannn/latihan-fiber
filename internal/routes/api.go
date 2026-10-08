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

	// 3-7. Students Routes
	students := api.Group("/students", middlewares.Protected())
	students.Get("/", middlewares.IsAdmin(), handlers.GetStudents)           // GET List (Admin)
	students.Post("/", middlewares.IsAdmin(), handlers.CreateStudent)        // POST (Admin)
	students.Get("/:id", handlers.GetStudentByID)                            // GET Detail (Admin/Self)
	students.Put("/:id", middlewares.IsAdmin(), handlers.UpdateStudent)      // PUT (Admin)
	students.Delete("/:id", middlewares.IsAdmin(), handlers.DeleteStudent)   // DELETE (Admin)
}
