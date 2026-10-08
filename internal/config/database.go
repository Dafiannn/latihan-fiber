package config

import (
	"log"
	"os"

	"github.com/dafian/siakad-mini/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Membaca koneksi langsung menggunakan Connection String (URI)
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("❌ DB_URL belum diatur di file .env")
	}

	// Karena kita menggunakan Connection Pooler (Supabase IPv4), kita harus mematikan fitur prepared statements bawaan GORM
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Wajib true saat pakai PgBouncer / Supabase Pooler
	}), &gorm.Config{})

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
