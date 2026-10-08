package handlers

import (
	"math"
	"strconv"

	"github.com/dafian/siakad-mini/internal/config"
	"github.com/dafian/siakad-mini/internal/models"
	"github.com/dafian/siakad-mini/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GET /api/v1/students (Hanya Admin)
func GetStudents(c *fiber.Ctx) error {
	db := config.DB

	// Parse Query Parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page <= 0 { page = 1 }

	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	if perPage <= 0 { perPage = 10 }
	if perPage > 50 { perPage = 50 } // Syarat PDF: maks 50

	prodi := c.Query("prodi")
	angkatan := c.Query("angkatan")
	search := c.Query("search")
	sort := c.Query("sort", "nama") // default sort

	query := db.Model(&models.Student{})

	// Filter
	if prodi != "" {
		query = query.Where("prodi = ?", prodi)
	}
	if angkatan != "" {
		query = query.Where("angkatan = ?", angkatan)
	}
	if search != "" {
		query = query.Where("nim LIKE ? OR nama ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Count Total Data for Meta Pagination
	var total int64
	query.Count(&total)

	// Sorting
	if sort == "-ipk_terakhir" {
		query = query.Order("ipk_terakhir desc")
	} else {
		query = query.Order("nama asc")
	}

	// Pagination
	offset := (page - 1) * perPage
	var students []models.Student
	query.Offset(offset).Limit(perPage).Find(&students)

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage == 0 { lastPage = 1 }

	meta := fiber.Map{
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    lastPage,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil daftar mahasiswa", students, meta)
}

type StudentInput struct {
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Email       string  `json:"email"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

// POST /api/v1/students (Hanya Admin)
func CreateStudent(c *fiber.Ctx) error {
	input := new(StudentInput)
	if err := c.BodyParser(input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Gagal memproses input", nil)
	}

	// Validasi Sederhana Sesuai PDF (422)
	errors := make(map[string][]string)
	if len(input.NIM) != 12 {
		errors["nim"] = []string{"NIM wajib unik 12 digit"}
	}
	if input.Nama == "" {
		errors["nama"] = []string{"Nama wajib diisi"}
	}
	if input.Email == "" {
		errors["email"] = []string{"Email wajib unik"}
	}
	if input.Prodi == "" {
		errors["prodi"] = []string{"Prodi wajib diisi"}
	}
	if input.Angkatan <= 0 {
		errors["angkatan"] = []string{"Angkatan wajib diisi dengan tahun (misal: 2024)"}
	}
	if len(errors) > 0 {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errors)
	}

	db := config.DB

	// Cek Duplikasi
	var count int64
	db.Model(&models.User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validasi gagal", fiber.Map{"email": []string{"Email sudah terdaftar"}})
	}
	db.Model(&models.Student{}).Where("nim = ?", input.NIM).Count(&count)
	if count > 0 {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validasi gagal", fiber.Map{"nim": []string{"NIM sudah terdaftar"}})
	}

	// Syarat PDF: Membuat record users & students dalam satu transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		hashedPassword := utils.HashPassword(input.NIM) // Password awal = NIM

		user := models.User{
			Email:    input.Email,
			Password: hashedPassword,
			Role:     "mahasiswa",
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		student := models.Student{
			UserID:      user.ID,
			NIM:         input.NIM,
			Nama:        input.Nama,
			Prodi:       input.Prodi,
			Angkatan:    input.Angkatan,
			IPKTerakhir: input.IPKTerakhir,
		}
		if err := tx.Create(&student).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat data mahasiswa", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Berhasil menambahkan mahasiswa", nil, nil)
}

// GET /api/v1/students/:id (Admin & Mahasiswa Sendiri)
func GetStudentByID(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	db := config.DB
	var student models.Student

	// Cari student (termasuk ngecek soft delete otomatis oleh GORM)
	if err := db.First(&student, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan", nil)
	}

	// Validasi Hak Akses (403): Mahasiswa hanya bisa lihat datanya sendiri
	if role == "mahasiswa" && student.UserID != userID {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak. Tidak dapat melihat data mahasiswa lain", nil)
	}

	// Hitung Total SKS yang diambil dari tabel Enrollments & Courses
	var totalSKS int
	db.Table("enrollments").
		Joins("JOIN courses ON enrollments.course_id = courses.id").
		Where("enrollments.student_id = ? AND enrollments.deleted_at IS NULL", student.ID). // asumsi deleted_at enrollment kosong
		Select("COALESCE(SUM(courses.sks), 0)").
		Scan(&totalSKS)

	// Hitung Batas SKS berdasar IPK (Business Rule 1)
	batasSKS := 18 // IPK < 2.50
	if student.IPKTerakhir >= 3.00 {
		batasSKS = 24
	} else if student.IPKTerakhir >= 2.50 {
		batasSKS = 21
	}

	data := fiber.Map{
		"student":   student,
		"total_sks": totalSKS,
		"batas_sks": batasSKS,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil detail mahasiswa", data, nil)
}

// PUT /api/v1/students/:id (Admin)
func UpdateStudent(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.DB

	var student models.Student
	if err := db.First(&student, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan", nil)
	}

	input := new(StudentInput)
	if err := c.BodyParser(input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Gagal memproses input", nil)
	}

	// NIM tidak boleh diubah, jadi hanya ubah field selain NIM
	student.Nama = input.Nama
	student.Prodi = input.Prodi
	student.Angkatan = input.Angkatan
	student.IPKTerakhir = input.IPKTerakhir

	if err := db.Save(&student).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui mahasiswa", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Berhasil memperbarui data mahasiswa", student, nil)
}

// DELETE /api/v1/students/:id (Admin)
func DeleteStudent(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.DB

	var student models.Student
	if err := db.First(&student, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan", nil)
	}

	// Soft Delete Student beserta User akun loginnya
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&student).Error; err != nil {
			return err
		}
		// Hapus juga akun usernya agar tidak bisa login (Syarat PDF: mahasiswa yg dihapus tidak dapat login)
		if err := tx.Where("id = ?", student.UserID).Delete(&models.User{}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus mahasiswa", nil)
	}

	return c.Status(fiber.StatusNoContent).SendString("") // 204 No Content
}
