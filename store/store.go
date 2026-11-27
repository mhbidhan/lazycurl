package store

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// InitDB opens the local SQLite database and migrates the collection schema.
// The file lives in tmp/ relative to the working directory.
func InitDB() *gorm.DB {
	if err := os.MkdirAll("tmp", 0o755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	db, err := gorm.Open(sqlite.Open("tmp/lazycurl.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Schema migrations
	if err := db.AutoMigrate(&Collection{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	return db
}
