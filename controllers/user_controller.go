package controllers

import (
	"kasir-backend/config"
	"kasir-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetPegawai: Mengambil seluruh daftar kasir
func GetPegawai(c *gin.Context) {
	var pegawai []models.User
	if err := config.DB.Where("role = ?", "kasir").Find(&pegawai).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pegawai"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pegawai})
}

// CreatePegawai: Owner menambahkan akun kasir baru
func CreatePegawai(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nama     string `json:"nama"`
		Name     string `json:"name"` // Cadangan jika frontend memakai key 'name'
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah: " + err.Error()})
		return
	}

	// Antisipasi jika frontend mengirim 'name' alih-alih 'nama'
	namaPegawai := input.Nama
	if namaPegawai == "" {
		namaPegawai = input.Name
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	pegawai := models.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Nama:     namaPegawai,
		Role:     "kasir",
	}

	if err := config.DB.Create(&pegawai).Error; err != nil {
		// Mengembalikan error Conflict agar terlihat jelas bahwa data bentrok
		c.JSON(http.StatusConflict, gin.H{"error": "Gagal! Username mungkin sudah digunakan oleh akun lain."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Akun pegawai berhasil ditambahkan",
		"data":    pegawai,
	})
}

// DeletePegawai: Owner menghapus akun kasir
func DeletePegawai(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus pegawai"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Akun pegawai berhasil dihapus"})
}
