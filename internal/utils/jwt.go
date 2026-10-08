package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken membuat JWT token yang berisi user_id dan role
func GenerateToken(userID uint, role string) (string, int64, error) {
	exp := time.Now().Add(time.Hour * 24).Unix() // Expired dalam 24 jam
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")

	t, err := token.SignedString([]byte(secret))
	return t, exp, err
}
