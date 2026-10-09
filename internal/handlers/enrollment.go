package handlers

import (
	"fmt"

	"github.com/dafian/siakad-mini/internal/config"
	"github.com/dafian/siakad-mini/internal/models"
	"github.com/dafian/siakad-mini/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EnrollmentInput struct {
	CourseID      uint   `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"` // cth: "2026/2027-Ganjil"
}

// POST /api/v1/enrollments (Hanya Mahasiswa)
func CreateEnrollment(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "mahasiswa" {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Hanya mahasiswa yang dapat mengambil KRS", nil)
	}

	input := new(EnrollmentInput)
	if err := c.BodyParser(input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validasi gagal", nil)
	}
	if input.CourseID == 0 || input.TahunAkademik == "" {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validasi gagal", fiber.Map{"error": "course_id dan tahun_akademik wajib diisi"})
	}

	db := config.DB
	userID := c.Locals("user_id").(uint)

	// Tarik data mahasiswa dari user_id yang sedang login
	var student models.Student
	if err := db.Where("user_id = ?", userID).First(&student).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Data profil mahasiswa tidak ditemukan", nil)
	}

	// Cek Duplikasi KRS (409)
	var existingCount int64
	db.Model(&models.Enrollment{}).Where("student_id = ? AND course_id = ? AND tahun_akademik = ?", student.ID, input.CourseID, input.TahunAkademik).Count(&existingCount)
	if existingCount > 0 {
		return utils.ErrorResponse(c, fiber.StatusConflict, "Mata kuliah tersebut sudah diambil pada tahun akademik ini", nil)
	}

	// Kalkulasi batas maksimal SKS (Business Rule 1)
	batasSKS := 18
	if student.IPKTerakhir >= 3.00 {
		batasSKS = 24
	} else if student.IPKTerakhir >= 2.50 {
		batasSKS = 21
	}

	// Gunakan Database Transaction dengan ROW LOCKING (Syarat dari PDF)
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Lock Row Tabel Courses untuk mencegah kebocoran Kuota
		var course models.Course
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&course, input.CourseID).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "Mata kuliah tidak ditemukan")
		}

		// 2. Hitung jumlah mahasiswa yang sudah mengambil (cek kuota)
		var terisi int64
		tx.Model(&models.Enrollment{}).Where("course_id = ?", course.ID).Count(&terisi)
		if terisi >= int64(course.Kuota) {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "Kuota mata kuliah sudah penuh")
		}

		// 3. Hitung SKS yang sudah diambil mahasiswa di tahun ini
		var totalSKSSaatIni int
		tx.Table("enrollments").
			Joins("JOIN courses ON enrollments.course_id = courses.id").
			Where("enrollments.student_id = ? AND enrollments.tahun_akademik = ?", student.ID, input.TahunAkademik).
			Select("COALESCE(SUM(courses.sks), 0)").Scan(&totalSKSSaatIni)

		// 4. Validasi Batas SKS
		if totalSKSSaatIni+course.SKS > batasSKS {
			sisa := batasSKS - totalSKSSaatIni
			msg := fmt.Sprintf("Total SKS melebihi batas. Batas SKS kamu: %d, Sisa SKS: %d, SKS Matkul ini: %d", batasSKS, sisa, course.SKS)
			return fiber.NewError(fiber.StatusUnprocessableEntity, msg)
		}

		// 5. Simpan ke database
		enrollment := models.Enrollment{
			StudentID:     student.ID,
			CourseID:      course.ID,
			TahunAkademik: input.TahunAkademik,
		}
		if err := tx.Create(&enrollment).Error; err != nil {
			return err
		}

		return nil
	})

	// Penanganan Error dari Transaction
	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			// Mengirim balik 422 atau 404
			return utils.ErrorResponse(c, e.Code, e.Message, nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil mata kuliah", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Berhasil mengambil mata kuliah (KRS)", nil, nil)
}

// DELETE /api/v1/enrollments/:id
func DeleteEnrollment(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "mahasiswa" {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Hanya mahasiswa yang dapat membatalkan KRS", nil)
	}

	enrollmentID := c.Params("id")
	userID := c.Locals("user_id").(uint)
	db := config.DB

	// Cari student yang sedang login
	var student models.Student
	if err := db.Where("user_id = ?", userID).First(&student).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan", nil)
	}

	var enrollment models.Enrollment
	if err := db.First(&enrollment, enrollmentID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Data KRS tidak ditemukan", nil)
	}

	// Hanya boleh membatalkan KRS miliknya sendiri (Syarat PDF: 403)
	if enrollment.StudentID != student.ID {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak. Ini bukan KRS milikmu", nil)
	}

	if err := db.Delete(&enrollment).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membatalkan KRS", nil)
	}

	// 204 No Content
	return c.Status(fiber.StatusNoContent).SendString("")
}
