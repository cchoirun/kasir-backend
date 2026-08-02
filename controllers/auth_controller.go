package controllers

import (
	"kasir-backend/config"
	"kasir-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Setup JWT Secret (Bisa disesuaikan nanti)
var jwtSecret = []byte("rahasia_kasir_qris")

// SeedOwner: Fungsi untuk membuat akun Owner default saat aplikasi pertama kali dijalankan
func SeedOwner() {
	var count int64
	config.DB.Model(&models.User{}).Where("role = ?", "owner").Count(&count)

	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("owner123"), bcrypt.DefaultCost)
		owner := models.User{
			Nama:     "Pemilik Toko",
			Username: "owner",
			Password: string(hashedPassword), // Menggunakan Password, bukan PasswordHash
			Role:     "owner",
		}
		config.DB.Create(&owner)
	}
}

// Login: Autentikasi User (Owner / Kasir)
func Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan password wajib diisi"})
		return
	}

	var user models.User
	if err := config.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username tidak ditemukan"})
		return
	}

	// Bandingkan password input dengan password di database
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password salah"})
		return
	}

	// Generate JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token otorisasi"})
		return
	}

	// Response sukses
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":       user.ID,
			"nama":     user.Nama,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// UpdateProfile: Mengubah nama, foto, no telepon, dan password
func UpdateProfile(c *gin.Context) {
	userIDClaim, _ := c.Get("user_id")
	userID := uint(userIDClaim.(float64))

	var input struct {
		Nama        string `json:"nama"`
		NoTelepon   string `json:"no_telepon"`
		FotoProfil  string `json:"foto_profil"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak lengkap"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	user.Nama = input.Nama
	user.NoTelepon = input.NoTelepon
	if input.FotoProfil != "" {
		user.FotoProfil = input.FotoProfil
	}

	// Jika user juga mengisi kolom password lama dan baru
	if input.OldPassword != "" && input.NewPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password lama salah"})
			return
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
		user.Password = string(hashedPassword)
	}

	config.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message": "Profil berhasil diperbarui",
		"user": gin.H{
			"id":          user.ID,
			"nama":        user.Nama,
			"username":    user.Username,
			"role":        user.Role,
			"no_telepon":  user.NoTelepon,
			"foto_profil": user.FotoProfil,
		},
	})
}
