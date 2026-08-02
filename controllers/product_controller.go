package controllers

import (
	"kasir-backend/config"
	"kasir-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductInput struct {
	Nama        string  `json:"nama" binding:"required"`
	Kategori    string  `json:"kategori"`
	Harga       float64 `json:"harga" binding:"required"`
	Stok        int     `json:"stok"`
	StokMinimum int     `json:"stok_minimum"`
	FotoURL     string  `json:"foto_url"`
}

// GetProducts: Mengambil daftar semua produk
func GetProducts(c *gin.Context) {
	var products []models.Product
	if err := config.DB.Where("status = ?", "active").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": products})
}

// CreateProduct: Menambah produk baru (Khusus Owner)
func CreateProduct(c *gin.Context) {
	var input struct {
		Nama        string  `json:"nama" binding:"required"`
		Kategori    string  `json:"kategori"`
		Harga       float64 `json:"harga" binding:"required"`
		Stok        int     `json:"stok" binding:"required"`
		StokMinimum int     `json:"stok_minimum"`
		FotoProduk  string  `json:"foto_produk"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah"})
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

	config.DB.Create(&product)
	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil ditambahkan", "product": product})
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
	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil diperbarui"})
}

// DeleteProduct: Menghapus/menonaktifkan produk (Khusus Owner)
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	// Soft delete / ubah status jadi inactive
	config.DB.Model(&product).Update("status", "inactive")

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil dinonaktifkan/dihapus"})
}
