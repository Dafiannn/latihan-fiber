package main

import (
	"log"

	"github.com/dafian/siakad-mini/internal/config"
	"github.com/dafian/siakad-mini/internal/routes"
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

	// Daftarkan semua routes
	routes.SetupRoutes(app)

	// Jalankan server di port 3000
	log.Fatal(app.Listen(":3000"))
}
