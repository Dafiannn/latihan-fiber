package config

import (
	"fmt"
	"log"
	"os"

	"github.com/dafian/siakad-mini/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Membaca konfigurasi dari file .env
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta", host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal terhubung ke Database: \n", err)
	}

	log.Println("✅ Berhasil terhubung ke Database PostgreSQL")

	// Menjalankan Auto Migrate untuk membuat tabel jika belum ada
	err = db.AutoMigrate(
		&models.User{},
		&models.Student{},
		&models.Course{},
		&models.Enrollment{},
	)
	if err != nil {
		log.Fatal("❌ Gagal menjalankan Auto Migrate: \n", err)
	}
	
	log.Println("✅ Auto Migrate Database berhasil (Tabel sudah siap)")

	DB = db
}
