package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(userID uint, role string, secret string) (string, error) {
	// Masukkan user_id DAN role ke dalam token claims
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
