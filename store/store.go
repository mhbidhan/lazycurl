package store

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	if err := os.MkdirAll("tmp", 0o755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	db, err := gorm.Open(sqlite.Open("tmp/lazycurl.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Schema migrations
	db.AutoMigrate(&Collection{})

	return db
}
