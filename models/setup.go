package models

import (
	"time"
)

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Nama       string    `json:"nama"`
	Username   string    `gorm:"unique" json:"username"`
	Password   string    `json:"-"`
	Role       string    `json:"role"`
	NoTelepon  string    `json:"no_telepon"`                   // BARU
	FotoProfil string    `gorm:"type:text" json:"foto_profil"` // BARU
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Nama        string    `json:"nama"`
	Kategori    string    `json:"kategori"`
	Harga       float64   `json:"harga"`
	Stok        int       `json:"stok"`
	StokMinimum int       `json:"stok_minimum"`
	FotoProduk  string    `gorm:"type:text" json:"foto_produk"`
	Status      string    `gorm:"type:varchar(20);default:'active'" json:"status"` // BARU: Kolom Status agar bisa di-soft-delete
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Transaction struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	KasirID         uint              `json:"kasir_id"`
	Kasir           User              `gorm:"foreignKey:KasirID" json:"kasir"`
	TotalAmount     float64           `gorm:"not null" json:"total_amount"`
	Discount        float64           `gorm:"default:0" json:"discount"`
	MetodeBayar     string            `gorm:"type:varchar(20);not null" json:"metode_bayar"`    // qris / tunai
	Status          string            `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending / lunas / batal
	QrisReferenceID string            `gorm:"type:varchar(100)" json:"qris_reference_id"`
	Items           []TransactionItem `gorm:"foreignKey:TransactionID" json:"items"`
	CreatedAt       time.Time         `json:"created_at"`
	PaidAt          *time.Time        `json:"paid_at"`
}

type TransactionItem struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	TransactionID uint    `json:"transaction_id"`
	ProductID     uint    `json:"product_id"`
	Product       Product `gorm:"foreignKey:ProductID" json:"product"`
	Qty           int     `gorm:"not null" json:"qty"`
	HargaSatuan   float64 `gorm:"not null" json:"harga_satuan"`
	Subtotal      float64 `gorm:"not null" json:"subtotal"`
}

type StockLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProductID  uint      `json:"product_id"`
	Product    Product   `gorm:"foreignKey:ProductID" json:"product"`
	Tipe       string    `gorm:"type:varchar(20);not null" json:"tipe"` // masuk / keluar
	Jumlah     int       `gorm:"not null" json:"jumlah"`
	Keterangan string    `gorm:"type:text" json:"keterangan"`
	CreatedAt  time.Time `json:"created_at"`
}
