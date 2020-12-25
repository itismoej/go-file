package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectAndMigrate() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	_ = db.AutoMigrate(&File{})
	return db
}
