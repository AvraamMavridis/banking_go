package db

import (
	"bank_api_go/entities"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(dbPath string) (*gorm.DB, error) {
	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := database.AutoMigrate(&entities.Account{}, &entities.IdempotencyRecord{}); err != nil {
		return nil, err
	}

	return database, nil
}
