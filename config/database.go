package config

import (
	"fmt"
	"kasir-backend/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default env variables")
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbname, port)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	database.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Transaction{},
		&models.TransactionItem{},
		&models.StockLog{},
	)

	// database.Migrator().DropTable(&models.User{}) // nonaktifkan kalau sudah direset

	DB = database
	fmt.Println("Database connection successful!")
}
