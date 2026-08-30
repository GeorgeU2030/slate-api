package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"slate-api/internal/config"
	"slate-api/internal/models"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Silent
	if !cfg.IsProduction() {
		logLevel = logger.Info 
	}

	db, err := gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("running auto migration...")
	return db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Title{},
		&models.UserTitle{},
	)
}