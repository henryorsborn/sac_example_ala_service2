package ala_service

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func OpenStore(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn),
		&gorm.Config{})
	if err != nil {
		return nil, err
	}

	// AutoMigrate creates the table if it doesn't 	exist, and adds missing columns.
	// It does NOT delete columns or change types — 		those require manual migrations.
	if err := db.AutoMigrate(&Alias{}); err != nil {
		return nil, err
	}

	log.Println("connected to postgres, aliases table ready")
	return db, nil
}
