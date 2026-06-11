package utils

import (
	"log"
	"os"
	"path/filepath"
	"wedding-system/config"
	"wedding-system/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dbDir := filepath.Dir(config.AppConfig.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(config.AppConfig.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.ServiceItem{},
		&models.Package{},
		&models.PackageItem{},
		&models.Customer{},
		&models.QuoteProposal{},
		&models.QuoteItem{},
		&models.Contract{},
		&models.Schedule{},
		&models.LuckyDay{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database initialized successfully")
}
