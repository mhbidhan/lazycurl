package store

import (
	"gorm.io/gorm"
)

type Collection struct {
	gorm.Model
	Name string `gorm:"not null"`
}

// NewCollection builds a collection ready to save.
func NewCollection(name string) *Collection {
	return &Collection{
		Name: name,
	}
}

// TableName names the table GORM maps collections to.
func (*Collection) TableName() string {
	return "collections"
}

// GetCollections returns every collection that has not been soft-deleted.
func GetCollections(db *gorm.DB) ([]Collection, error) {
	var colls []Collection
	err := db.Table("collections").Find(&colls).Error
	return colls, err
}

// SaveCollection inserts or updates a collection, assigning an ID on insert.
func SaveCollection(db *gorm.DB, coll *Collection) error {
	return db.Table(coll.TableName()).Save(coll).Error
}

// DeleteCollection soft deletes by primary key, so rows stay recoverable and
// GetCollections stops returning them.
func DeleteCollection(db *gorm.DB, id uint) error {
	return db.Delete(&Collection{}, id).Error
}
