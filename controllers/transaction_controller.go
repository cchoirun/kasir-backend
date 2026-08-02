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
	// Karena logika transaksi cukup kompleks (mengurangi stok, menyimpan item, dll),
	// ini adalah kerangka dasar agar bisa di-compile.
	// Jika kamu sebelumnya sudah punya kode CreateTransaction yang lebih lengkap, gunakan milikmu.
	var input models.Transaction
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah"})
		return
	}

	input.Tanggal = time.Now()
	config.DB.Create(&input)

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaksi berhasil disimpan",
		"data":    input,
	})
}

// GetTransactions: Mengambil riwayat transaksi
func GetTransactions(c *gin.Context) {
	var transactions []models.Transaction
	// Gunakan Preload jika ada relasi ke tabel TransactionItem
	if err := config.DB.Order("created_at desc").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat riwayat transaksi"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

// GetQrisSimulation: Simulasi pengecekan pembayaran QRIS
func GetQrisSimulation(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message":        "Simulasi QRIS sukses",
		"transaction_id": id,
		"status":         "paid",
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
		"omzet_hari_ini":  0, // Bisa disesuaikan dengan query SUM(total)
		"total_transaksi": totalTransaksi,
		"produk_terlaris": []string{},
		"message":         "Data analitik siap disambungkan",
	})
}
