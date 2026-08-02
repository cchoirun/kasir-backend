package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(getSecretKey())

func getSecretKey() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "rahasia_default_kasir_qris" // fallback key
	}
	return secret
}

type Claims struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(id uint, username, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token berlaku 24 jam
	claims := &Claims{
		ID:       id,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, err
}
