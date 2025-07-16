package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"assignment1/models"
)

var DB *gorm.DB

func Connect(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to the database : ", err)
	}

	if err := DB.AutoMigrate(&models.Rate{}); err != nil {
		log.Fatal("Auto migration failed : ", err)
	}
}
