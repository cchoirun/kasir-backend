package controllers

import (
	"fmt"
	"kasir-backend/config"
	"kasir-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ItemInput struct {
	ProductID uint `json:"product_id" binding:"required"`
	Qty       int  `json:"qty" binding:"required"`
}

type TransactionInput struct {
	MetodeBayar string      `json:"metode_bayar" binding:"required"` // qris / tunai
	Discount    float64     `json:"discount"`
	Items       []ItemInput `json:"items" binding:"required"`
}

// CreateTransaction: Membuat transaksi baru
func CreateTransaction(c *gin.Context) {
	var input TransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(float64) // dari middleware JWT

	var totalAmount float64 = 0
	var transactionItems []models.TransactionItem

	// Mulai Database Transaction (GORM Transaction)
	tx := config.DB.Begin()

	for _, itemInput := range input.Items {
		var product models.Product
		if err := tx.First(&product, itemInput.ProductID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Produk dengan ID %d tidak ditemukan", itemInput.ProductID)})
			return
		}

		// Cek stok apakah mencukupi (kecuali untuk pembayaran tunai/qris, stok dikurangi saat lunas atau langsung)
		if product.Stok < itemInput.Qty {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Stok produk '%s' tidak mencukupi (sisa: %d)", product.Nama, product.Stok)})
			return
		}

		subtotal := product.Harga * float64(itemInput.Qty)
		totalAmount += subtotal

		transactionItems = append(transactionItems, models.TransactionItem{
			ProductID:   product.ID,
			Qty:         itemInput.Qty,
			HargaSatuan: product.Harga,
			Subtotal:    subtotal,
		})

		// Jika metode bayar tunai, langsung kurangi stok produk sekarang
		if input.MetodeBayar == "tunai" {
			product.Stok -= itemInput.Qty
			tx.Save(&product)

			// Catat log stok keluar
			tx.Create(&models.StockLog{
				ProductID:  product.ID,
				Tipe:       "keluar",
				Jumlah:     itemInput.Qty,
				Keterangan: "Penjualan Tunai",
			})
		}
	}

	finalAmount := totalAmount - input.Discount
	if finalAmount < 0 {
		finalAmount = 0
	}

	status := "pending"
	if input.MetodeBayar == "tunai" {
		status = "lunas"
	}

	now := time.Now()
	var paidAt *time.Time = nil
	if status == "lunas" {
		paidAt = &now
	}

	transaction := models.Transaction{
		KasirID:         uint(userID),
		TotalAmount:     finalAmount,
		Discount:        input.Discount,
		MetodeBayar:     input.MetodeBayar,
		Status:          status,
		QrisReferenceID: fmt.Sprintf("QRIS-DUMMY-%d", time.Now().UnixNano()),
		Items:           transactionItems,
		PaidAt:          paidAt,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan transaksi"})
		return
	}

	tx.Commit()

	responseMessage := "Transaksi tunai berhasil disimpan"
	if input.MetodeBayar == "qris" {
		responseMessage = "Transaksi QRIS dibuat, silakan generate QR code"
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     responseMessage,
		"transaction": transaction,
	})
}

func GetTransactions(c *gin.Context) {
	var transactions []models.Transaction

	// Mengambil data transaksi beserta relasi item dan produknya, diurutkan dari yang terbaru
	if err := config.DB.Preload("Items.Product").Preload("Kasir").Order("created_at desc").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil riwayat transaksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// GenerateQrisDummy: Simulasi generate QRIS dinamis via Payment Gateway (Dummy)
func GenerateQrisDummy(c *gin.Context) {
	id := c.Param("id")
	var transaction models.Transaction

	if err := config.DB.Preload("Items.Product").First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	if transaction.MetodeBayar != "qris" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaksi ini bukan metode pembayaran QRIS"})
		return
	}

	// Mock/Dummy QRIS Response ala Xendit/Midtrans
	qrString := fmt.Sprintf("00020101021226580011ID.CO.DUMMY.WWW011893600914000000000303U015204000053033605802ID5910KASIR_QRIS6007SURABAYA6304%s", transaction.QrisReferenceID)

	c.JSON(http.StatusOK, gin.H{
		"message":            "QRIS berhasil digenerate (Dummy)",
		"transaction_id":     transaction.ID,
		"total_amount":       transaction.TotalAmount,
		"qris_reference_id":  transaction.QrisReferenceID,
		"qr_string":          qrString,
		"qr_image_url":       fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s", qrString),
		"status":             transaction.Status,
		"expired_in_minutes": 10,
	})
}

// SimulateWebhookPayment: Simulasi webhook dari Payment Gateway saat pelanggan sukses bayar QRIS
func SimulateWebhookPayment(c *gin.Context) {
	var input struct {
		QrisReferenceID string `json:"qris_reference_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var transaction models.Transaction
	if err := config.DB.Preload("Items").Where("qris_reference_id = ?", input.QrisReferenceID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referensi QRIS tidak ditemukan"})
		return
	}

	if transaction.Status == "lunas" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaksi sudah dibayar sebelumnya"})
		return
	}

	// Mulai Database Transaction untuk update status & kurangi stok
	tx := config.DB.Begin()

	now := time.Now()
	transaction.Status = "lunas"
	transaction.PaidAt = &now
	tx.Save(&transaction)

	// Kurangi stok produk karena pembayaran QRIS baru dinyatakan lunas sekarang
	for _, item := range transaction.Items {
		var product models.Product
		tx.First(&product, item.ProductID)
		product.Stok -= item.Qty
		tx.Save(&product)

		// Catat log stok keluar
		tx.Create(&models.StockLog{
			ProductID:  product.ID,
			Tipe:       "keluar",
			Jumlah:     item.Qty,
			Keterangan: fmt.Sprintf("Penjualan QRIS (Trx ID: %d)", transaction.ID),
		})
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":     "Webhook diterima: Pembayaran QRIS Berhasil (Lunas)",
		"transaction": transaction,
	})
}

// GetDashboardStats: Mengambil ringkasan omzet, total transaksi, dan produk terlaris
func GetDashboardStats(c *gin.Context) {
	var totalOmzet float64
	var totalTransaksi int64

	// Hitung total omzet
	config.DB.Model(&models.Transaction{}).Where("status = ?", "lunas").Select("COALESCE(SUM(total_amount), 0)").Scan(&totalOmzet)

	// Hitung jumlah transaksi lunas
	config.DB.Model(&models.Transaction{}).Where("status = ?", "lunas").Count(&totalTransaksi)

	// Hitung 5 Produk Terlaris menggunakan Raw SQL
	type TopProduct struct {
		Nama       string  `json:"nama"`
		Terjual    int     `json:"terjual"`
		Pendapatan float64 `json:"pendapatan"`
	}
	var topProducts []TopProduct

	query := `
		SELECT p.nama, SUM(ti.qty) as terjual, SUM(ti.subtotal) as pendapatan
		FROM transaction_items ti
		JOIN transactions t ON t.id = ti.transaction_id
		JOIN products p ON p.id = ti.product_id
		WHERE t.status = 'lunas'
		GROUP BY p.id, p.nama
		ORDER BY terjual DESC
		LIMIT 5
	`
	config.DB.Raw(query).Scan(&topProducts)

	c.JSON(http.StatusOK, gin.H{
		"total_omzet":     totalOmzet,
		"total_transaksi": totalTransaksi,
		"top_products":    topProducts,
	})
}
