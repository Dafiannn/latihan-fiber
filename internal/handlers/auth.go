package handlers

import (
	"sync"
	"time"

	"github.com/dafian/siakad-mini/internal/config"
	"github.com/dafian/siakad-mini/internal/models"
	"github.com/dafian/siakad-mini/internal/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Mekanisme Rate Limiter in-memory sederhana untuk mencatat gagal login per IP
var (
	failedLogins = make(map[string]int)
	loginMutex   sync.Mutex
)

// Reset rate limiter setiap 1 menit (karena aturan PDF: 5 kali per menit)
func init() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			loginMutex.Lock()
			failedLogins = make(map[string]int)
			loginMutex.Unlock()
		}
	}()
}

func recordFailedLogin(ip string) {
	loginMutex.Lock()
	failedLogins[ip]++
	loginMutex.Unlock()
}

// POST /api/v1/auth/login
func Login(c *fiber.Ctx) error {
	ip := c.IP()
	
	// Cek Rate Limit (429)
	loginMutex.Lock()
	attempts := failedLogins[ip]
	loginMutex.Unlock()
	if attempts >= 5 {
		return utils.ErrorResponse(c, fiber.StatusTooManyRequests, "Terlalu banyak percobaan. Gagal login lebih dari 5 kali per menit", nil)
	}

	// Parse Body
	input := new(LoginInput)
	if err := c.BodyParser(input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validasi gagal", fiber.Map{"body": "Format JSON tidak sesuai"})
	}

	// Validasi input manual (Sesuai PDF: 422)
	errors := make(map[string][]string)
	if input.Email == "" {
		errors["email"] = []string{"Email wajib diisi"}
	}
	if len(input.Password) < 8 {
		errors["password"] = []string{"Password wajib diisi dan minimal 8 karakter"}
	}
	if len(errors) > 0 {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errors)
	}

	// Cari User di DB (Sesuai PDF: 401 jika salah)
	db := config.DB
	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		recordFailedLogin(ip)
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Kredensial email atau password salah", nil)
	}

	// Bandingkan Password Hash (Sesuai PDF: 401 jika salah)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		recordFailedLogin(ip)
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Kredensial email atau password salah", nil)
	}

	// Generate Token JWT
	token, exp, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memproses token login", nil)
	}

	// Sesuai PDF: Status 200, return token + user info
	return utils.SuccessResponse(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   exp,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	}, nil)
}

// GET /api/v1/auth/me
func Me(c *fiber.Ctx) error {
	// Ambil data token dari Fiber Locals yang di-set oleh Middleware
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	db := config.DB
	var user models.User

	// Jika role mahasiswa, harus mengambil data tabel Students (Preload)
	if role == "mahasiswa" {
		if err := db.Preload("Student").First(&user, userID).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Data pengguna tidak ditemukan", nil)
		}
		
		return utils.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil profil mahasiswa", fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
			"student": fiber.Map{
				"nim":      user.Student.NIM,
				"nama":     user.Student.Nama,
				"prodi":    user.Student.Prodi,
				"angkatan": user.Student.Angkatan,
			},
		}, nil)
	}

	// Jika role admin (tidak butuh data student)
	if err := db.First(&user, userID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Data pengguna tidak ditemukan", nil)
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil profil admin", fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}, nil)
}
