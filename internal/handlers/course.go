package handlers

import (
	"github.com/dafian/siakad-mini/internal/config"
	"github.com/dafian/siakad-mini/internal/models"
	"github.com/dafian/siakad-mini/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type CourseResult struct {
	models.Course
	Terisi    int `json:"terisi"`
	SisaKuota int `json:"sisa_kuota"`
}

// GET /api/v1/courses
func GetCourses(c *fiber.Ctx) error {
	db := config.DB
	semester := c.Query("semester")
	search := c.Query("search")
	available := c.Query("available") == "true" // true/false string ke bool

	// Query dengan Join untuk menghitung Sisa Kuota
	query := db.Table("courses").
		Select("courses.*, (courses.kuota - COUNT(enrollments.id)) as sisa_kuota, COUNT(enrollments.id) as terisi").
		Joins("LEFT JOIN enrollments ON enrollments.course_id = courses.id").
		Group("courses.id")

	// Filter
	if semester != "" {
		query = query.Where("courses.semester = ?", semester)
	}
	if search != "" {
		query = query.Where("courses.kode_mk LIKE ? OR courses.nama_mk ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if available {
		query = query.Having("(courses.kuota - COUNT(enrollments.id)) > 0")
	}

	var results []CourseResult
	if err := query.Find(&results).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data mata kuliah", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data mata kuliah", results, nil)
}
