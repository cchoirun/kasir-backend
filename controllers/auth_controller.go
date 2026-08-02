package controllers

import (
	"kasir-backend/config"
	"kasir-backend/models"
	"kasir-backend/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// SeedOwner: Membuat akun owner default saat aplikasi pertama kali dijalankan
func SeedOwner() {
	var count int64
	config.DB.Model(&models.User{}).Where("role = ?", "owner").Count(&count)
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("owner123"), bcrypt.DefaultCost)
		owner := models.User{
			Username: "owner",
			Password: string(hashedPassword),
			Role:     "owner",
			Nama:     "Owner Toko",
		}
		config.DB.Create(&owner)
	}
}

// Login: Proses autentikasi user
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "rahasia_kasir_qris"
	}

	token, err := utils.GenerateToken(user.ID, user.Role, secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"nama":     user.Nama,
			"role":     user.Role,
		},
	})
}

// UpdateProfile: Mengubah nama user
func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		Nama string `json:"nama"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	user.Nama = input.Nama
	config.DB.Save(&user)

	// HARUS mengembalikan data user agar frontend tidak menyimpan "undefined"
	c.JSON(http.StatusOK, gin.H{
		"message": "Profil berhasil diperbarui",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"nama":     user.Nama,
			"role":     user.Role,
		},
	})
}

// ChangePassword: Mengubah kata sandi
func ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	// Cek password lama
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password lama salah"})
		return
	}

	// Hash password baru
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	config.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}
