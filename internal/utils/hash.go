package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword mengenkripsi string menjadi bcrypt hash
func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}
