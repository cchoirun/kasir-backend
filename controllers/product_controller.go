package controllers

import (
	"kasir-backend/config"
	"kasir-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProducts: Mengambil daftar semua produk
func GetProducts(c *gin.Context) {
	var products []models.Product

	// Dihapus pengecekan "status" agar tidak memicu Error 500 dari Database.
	// Kita cukup mengambil semua data produk yang ada.
	if err := config.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}

	// Frontend Vercel (Axios) umumnya mencari property "data", jadi kita ubah key-nya
	c.JSON(http.StatusOK, gin.H{
		"message": "Data produk berhasil dimuat",
		"data":    products,
	})
}

// CreateProduct: Menambah produk baru
func CreateProduct(c *gin.Context) {
	// Hapus binding:"required" untuk mencegah error input 0
	var input struct {
		Nama        string  `json:"nama"`
		Kategori    string  `json:"kategori"`
		Harga       float64 `json:"harga"`
		Stok        int     `json:"stok"`
		StokMinimum int     `json:"stok_minimum"`
		FotoProduk  string  `json:"foto_produk"`
	}

	// Kembalikan pesan err.Error() agar detail kesalahannya terbaca di frontend
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data ditolak: " + err.Error()})
		return
	}

	product := models.Product{
		Nama:        input.Nama,
		Kategori:    input.Kategori,
		Harga:       input.Harga,
		Stok:        input.Stok,
		StokMinimum: input.StokMinimum,
		FotoProduk:  input.FotoProduk,
	}

	if err := config.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan produk ke database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Produk berhasil ditambahkan",
		"data":    product,
	})
}

// UpdateProduct: Mengedit data stok dan detail produk
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Nama        string  `json:"nama"`
		Kategori    string  `json:"kategori"`
		Harga       float64 `json:"harga"`
		Stok        int     `json:"stok"`
		StokMinimum int     `json:"stok_minimum"`
		FotoProduk  string  `json:"foto_produk"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah"})
		return
	}

	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	product.Nama = input.Nama
	product.Kategori = input.Kategori
	product.Harga = input.Harga
	product.Stok = input.Stok
	product.StokMinimum = input.StokMinimum

	// Hanya update foto jika ada foto baru yang dikirim
	if input.FotoProduk != "" {
		product.FotoProduk = input.FotoProduk
	}

	config.DB.Save(&product)
	c.JSON(http.StatusOK, gin.H{
		"message": "Produk berhasil diperbarui",
		"data":    product,
	})
}

// DeleteProduct: Menghapus produk (Khusus Owner)
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	// Menghapus data secara langsung (Hard Delete).
	// (Atau GORM akan otomatis melakukan Soft Delete jika tabelmu menggunakan gorm.Model)
	if err := config.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus produk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil dihapus"})
}
