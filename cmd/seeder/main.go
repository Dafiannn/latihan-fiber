package main

import (
	"fmt"
	"log"

	"github.com/dafian/siakad-mini/internal/config"
	"github.com/dafian/siakad-mini/internal/models"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengenkripsi password menggunakan bcrypt
func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

func main() {
	// 1. Load env dan konek DB
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}
	config.ConnectDB()
	db := config.DB

	log.Println("Memulai proses seeding data...")

	// 2. Buat 1 Admin (jika belum ada)
	adminPassword := HashPassword("admin123")
	admin := models.User{
		Email:    "admin@siakad.com",
		Password: adminPassword,
		Role:     "admin",
	}
	// FirstOrCreate akan mencegah error duplikasi jika script ini dijalankan 2 kali
	db.Where(models.User{Email: "admin@siakad.com"}).FirstOrCreate(&admin)

	// 3. Buat 20 Mahasiswa
	for i := 1; i <= 20; i++ {
		nim := fmt.Sprintf("18722100%04d", i) // Format NIM 12 digit
		email := fmt.Sprintf("student%d@siakad.com", i)
		hashedPassword := HashPassword(nim) // Syarat PDF: password awal mahasiswa = NIM

		user := models.User{
			Email:    email,
			Password: hashedPassword,
			Role:     "mahasiswa",
		}
		
		var existingUser models.User
		if err := db.Where("email = ?", email).First(&existingUser).Error; err != nil {
			// Jika belum ada, buat User
			db.Create(&user)
			// Buat profil Student yang berelasi dengan User di atas
			student := models.Student{
				UserID:      user.ID,
				NIM:         nim,
				Nama:        fmt.Sprintf("Mahasiswa %d", i),
				Prodi:       "Sistem Informasi",
				Angkatan:    2023,
				IPKTerakhir: 3.50, // IPK dummy
			}
			db.Create(&student)
		}
	}

	// 4. Buat 10 Mata Kuliah
	for i := 1; i <= 10; i++ {
		kode := fmt.Sprintf("MK%03d", i)
		course := models.Course{
			KodeMK:   kode,
			NamaMK:   fmt.Sprintf("Mata Kuliah %d", i),
			SKS:      3,
			Semester: 3,
			Kuota:    40,
		}
		db.Where(models.Course{KodeMK: kode}).FirstOrCreate(&course)
	}

	log.Println("✅ Seeding selesai! Berhasil menambahkan 1 Admin, 20 Mahasiswa, 10 Mata Kuliah.")
}
