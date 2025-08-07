package database

import (
	"fmt"
	"log"
	
	"gorm.io/gorm"
	
	"turnate/internal/models"
)

func AutoMigrateModels(db *gorm.DB) error {
	if err := models.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	if err := models.CreateIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Auto-migration completed successfully")
	return nil
}