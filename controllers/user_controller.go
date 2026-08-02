package controllers

import (
	"kasir-backend/config"
	"kasir-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetUsers: Mengambil daftar semua kasir
func GetUsers(c *gin.Context) {
	var users []models.User
	// Hanya ambil data akun yang memiliki role 'kasir'
	config.DB.Where("role = ?", "kasir").Find(&users)
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// CreateUser: Mendaftarkan kasir baru
func CreateUser(c *gin.Context) {
	var input struct {
		Nama     string `json:"nama" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak lengkap"})
		return
	}

	// Cek apakah username sudah ada
	var existingUser models.User
	if err := config.DB.Where("username = ?", input.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username sudah digunakan!"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	user := models.User{
		Nama:     input.Nama,
		Username: input.Username,
		Password: string(hashedPassword),
		Role:     "kasir",
	}

	// PERBAIKAN: Tangkap error jika gagal menyimpan ke database
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan ke DB: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kasir berhasil ditambahkan"})
}

// DeleteUser: Menghapus akun kasir
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kasir"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Akun kasir berhasil dihapus"})
}
