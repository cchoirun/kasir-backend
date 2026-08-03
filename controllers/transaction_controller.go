package controllers

import (
	"kasir-backend/config"
	"kasir-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateTransaction: Membuat transaksi baru
func CreateTransaction(c *gin.Context) {
	var input models.Transaction

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah: " + err.Error()})
		return
	}

	if input.MetodeBayar == "tunai" {
		input.Status = "lunas"
		now := time.Now()
		input.PaidAt = &now
	} else {
		input.Status = "pending"
	}

	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan transaksi ke database: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaksi berhasil disimpan",
		"data":    input,
	})
}

// GetTransactions: Mengambil riwayat transaksi
func GetTransactions(c *gin.Context) {
	var transactions []models.Transaction

	err := config.DB.
		Preload("Kasir").
		Preload("Items").
		Preload("Items.Product").
		Order("created_at desc").
		Find(&transactions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat riwayat transaksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Riwayat transaksi berhasil dimuat",
		"data":    transactions,
	})
}

// GetQrisSimulation: Simulasi pengecekan pembayaran QRIS
func GetQrisSimulation(c *gin.Context) {
	id := c.Param("id")
	var transaction models.Transaction

	if err := config.DB.First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	// Generate QRIS Reference ID
	refID := "QRIS-SIM-" + id

	// Simpan Reference ID ke database agar bisa dicocokkan saat Webhook dipanggil
	config.DB.Model(&transaction).Update("qris_reference_id", refID)

	c.JSON(http.StatusOK, gin.H{
		"message":           "Simulasi QRIS sukses",
		"transaction_id":    transaction.ID,
		"qris_reference_id": refID,
		"total_amount":      transaction.TotalAmount,
		"qr_image_url":      "https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=" + refID,
		"status":            "pending",
	})
}

// WebhookPayment: Endpoint menerima notifikasi pelunasan QRIS
func WebhookPayment(c *gin.Context) {
	var input struct {
		QrisReferenceID string `json:"qris_reference_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah"})
		return
	}

	// Cari transaksi berdasarkan QRIS Reference ID
	var transaction models.Transaction
	if err := config.DB.Where("qris_reference_id = ?", input.QrisReferenceID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	// LOGIKA BARU: Update status transaksi menjadi lunas
	now := time.Now()
	config.DB.Model(&transaction).Updates(models.Transaction{
		Status: "lunas",
		PaidAt: &now,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Pembayaran QRIS berhasil dikonfirmasi dan dilunaskan",
	})
}

// GetDashboardAnalytics: Menampilkan ringkasan omzet dan transaksi di Dashboard Owner
func GetDashboardAnalytics(c *gin.Context) {
	var totalTransaksi int64
	var totalOmzet float64

	config.DB.Model(&models.Transaction{}).Where("status = ?", "lunas").Count(&totalTransaksi)

	config.DB.Table("transactions").Where("status = ?", "lunas").Select("COALESCE(SUM(total_amount), 0)").Scan(&totalOmzet)

	type TopProduct struct {
		Nama    string `json:"nama"`
		Terjual int    `json:"terjual"`
	}
	var topProducts []TopProduct

	config.DB.Table("transaction_items").
		Select("products.nama, SUM(transaction_items.qty) as terjual").
		Joins("JOIN products ON products.id = transaction_items.product_id").
		Joins("JOIN transactions ON transactions.id = transaction_items.transaction_id").
		Where("transactions.status = ?", "lunas").
		Group("products.id, products.nama").
		Order("terjual DESC").
		Limit(5).
		Scan(&topProducts)

	c.JSON(http.StatusOK, gin.H{
		"total_omzet":     totalOmzet,
		"total_transaksi": totalTransaksi,
		"top_products":    topProducts,
		"message":         "Data analitik berhasil dimuat",
	})
}
