package controllers

import (
	"kasir-backend/config"
	"kasir-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateTransaction: Membuat transaksi baru
func CreateTransaction(c *gin.Context) {
	var input models.Transaction

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah: " + err.Error()})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"message":           "Simulasi QRIS sukses",
		"transaction_id":    transaction.ID,
		"qris_reference_id": "QRIS-SIM-" + id,
		"total_amount":      transaction.TotalAmount,
		"qr_image_url":      "https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=SIMULASI-QRIS-" + id,
		"status":            "pending",
	})
}

// WebhookPayment: Endpoint untuk menerima notifikasi dari payment gateway asli nantinya
func WebhookPayment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook pembayaran berhasil diterima",
	})
}

// GetDashboardAnalytics: Menampilkan ringkasan omzet dan transaksi di Dashboard Owner
func GetDashboardAnalytics(c *gin.Context) {
	var totalTransaksi int64
	config.DB.Model(&models.Transaction{}).Count(&totalTransaksi)

	c.JSON(http.StatusOK, gin.H{
		"omzet_hari_ini":  0,
		"total_transaksi": totalTransaksi,
		"produk_terlaris": []string{},
		"message":         "Data analitik siap disambungkan",
	})
}
