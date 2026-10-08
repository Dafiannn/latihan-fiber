package main

import (
	"log"

	"github.com/dafian/siakad-mini/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables dari file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using default env vars")
	}

	// Inisialisasi Database (beserta Auto Migrate)
	config.ConnectDB()

	// Inisialisasi Fiber app
	app := fiber.New()

	// Route sederhana untuk test server
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to SIAKAD Mini API",
		})
	})

	// Jalankan server di port 3000
	log.Fatal(app.Listen(":3000"))
}
