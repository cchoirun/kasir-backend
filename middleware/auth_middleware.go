package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// PERHATIAN: Kunci ini harus sama persis dengan yang ada di auth_controller.go
var jwtSecret = []byte("rahasia_kasir_qris")

func AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak ditemukan"})
			c.Abort()
			return
		}

		// Menghapus kata "Bearer " dari string token
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Mem-parsing dan memvalidasi token menggunakan kunci rahasia
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau kadaluarsa"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal membaca data dari token"})
			c.Abort()
			return
		}

		userRole := claims["role"].(string)

		// Cek otorisasi peran (role)
		if requiredRole != "" && userRole != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki akses ke fitur ini"})
			c.Abort()
			return
		}

		// Lolos verifikasi, lanjutkan ke controller berikutnya
		c.Set("user_id", claims["user_id"])
		c.Set("role", userRole)
		c.Next()
	}
}
