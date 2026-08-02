package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware sekarang menerima satu argumen string (biasanya untuk membatasi role "owner" atau "kasir")
func AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Header otorisasi tidak ditemukan"})
			c.Abort()
			return
		}

		// Memisahkan "Bearer" dan tokennya
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Format token tidak valid"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "rahasia_kasir_qris"
		}

		// Membaca dan memverifikasi token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode enkripsi tidak valid")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau kadaluarsa"})
			c.Abort()
			return
		}

		// Mengambil data dari payload token dan menyimpannya ke context
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			userRole := fmt.Sprintf("%v", claims["role"])

			// Proteksi tambahan: Jika rute membutuhkan role tertentu ("owner" / "kasir")
			// kita blokir jika token yang dikirim memiliki role yang berbeda.
			if requiredRole == "owner" || requiredRole == "kasir" {
				if userRole != requiredRole {
					c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda tidak memiliki izin untuk halaman ini"})
					c.Abort()
					return
				}
			}

			c.Set("user_id", claims["user_id"])
			c.Set("role", userRole)
		}

		c.Next()
	}
}
