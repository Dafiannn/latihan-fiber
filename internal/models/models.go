package models

import (
	"time"

	"gorm.io/gorm"
)

// User merepresentasikan akun untuk login
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"` // Password disembunyikan di JSON response
	Role      string         `gorm:"type:varchar(20);not null" json:"role"` // 'admin' atau 'mahasiswa'
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Student *Student `gorm:"foreignKey:UserID" json:"student,omitempty"`
}

// Student merepresentasikan data profil mahasiswa
type Student struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;uniqueIndex" json:"user_id"` // Relasi 1-1 ke tabel users
	NIM         string         `gorm:"type:varchar(12);uniqueIndex;not null" json:"nim"`
	Nama        string         `gorm:"not null" json:"nama"`
	Prodi       string         `gorm:"not null" json:"prodi"`
	Angkatan    int            `gorm:"not null" json:"angkatan"`
	IPKTerakhir float64        `gorm:"type:numeric(3,2);default:0.00" json:"ipk_terakhir"` // Maks 4.00
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // Sesuai syarat soft delete

	Enrollments []Enrollment `gorm:"foreignKey:StudentID" json:"enrollments,omitempty"`
}

// Course merepresentasikan data mata kuliah
type Course struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	KodeMK    string         `gorm:"uniqueIndex;not null" json:"kode_mk"`
	NamaMK    string         `gorm:"not null" json:"nama_mk"`
	SKS       int            `gorm:"not null" json:"sks"`
	Semester  int            `gorm:"not null" json:"semester"`
	Kuota     int            `gorm:"not null" json:"kuota"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`

	Enrollments []Enrollment `gorm:"foreignKey:CourseID" json:"-"`
}

// Enrollment merepresentasikan pengambilan KRS mahasiswa
type Enrollment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	StudentID     uint      `gorm:"uniqueIndex:idx_student_course_year;not null" json:"student_id"`
	CourseID      uint      `gorm:"uniqueIndex:idx_student_course_year;not null" json:"course_id"`
	TahunAkademik string    `gorm:"type:varchar(20);uniqueIndex:idx_student_course_year;not null" json:"tahun_akademik"` 
	CreatedAt     time.Time `json:"created_at"`
}
