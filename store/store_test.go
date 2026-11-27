package store

import (
	"os"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Collection{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestInitDB(t *testing.T) {
	t.Chdir(t.TempDir())

	db := InitDB()
	if db == nil {
		t.Fatal("InitDB returned nil")
	}
	if _, err := os.Stat("tmp/lazycurl.db"); err != nil {
		t.Fatalf("database file missing: %v", err)
	}
	if !db.Migrator().HasTable(&Collection{}) {
		t.Fatal("collections table was not migrated")
	}
}

func TestCollectionCRUD(t *testing.T) {
	db := testDB(t)

	created := NewCollection("work")
	if created.Name != "work" {
		t.Fatalf("name = %q", created.Name)
	}
	if created.TableName() != "collections" {
		t.Fatalf("table = %q", created.TableName())
	}

	if err := SaveCollection(db, created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Fatal("save did not assign an ID")
	}

	rows, err := GetCollections(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "work" {
		t.Fatalf("rows = %+v", rows)
	}

	rows[0].Name = "home"
	if err := SaveCollection(db, &rows[0]); err != nil {
		t.Fatal(err)
	}
	rows, err = GetCollections(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "home" {
		t.Fatalf("rows after update = %+v", rows)
	}

	if err := DeleteCollection(db, created.ID); err != nil {
		t.Fatal(err)
	}
	rows, err = GetCollections(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("deleted collection still listed: %+v", rows)
	}

	var kept int64
	if err := db.Unscoped().Model(&Collection{}).Where("id = ?", created.ID).Count(&kept).Error; err != nil {
		t.Fatal(err)
	}
	if kept != 1 {
		t.Fatalf("soft-deleted row count = %d, want 1", kept)
	}
}

func TestCollectionErrors(t *testing.T) {
	db := testDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := GetCollections(db); err == nil {
		t.Error("GetCollections should report a closed database")
	}
	if err := SaveCollection(db, NewCollection("work")); err == nil {
		t.Error("SaveCollection should report a closed database")
	}
	if err := DeleteCollection(db, 1); err == nil {
		t.Error("DeleteCollection should report a closed database")
	}
}
