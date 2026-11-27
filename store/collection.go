package store

import (
	"gorm.io/gorm"
)

type Collection struct {
	gorm.Model
	Name string `gorm:"not null"`
}

func NewCollection(name string) *Collection {
	return &Collection{
		Name: name,
	}
}

func (*Collection) TableName() string {
	return "collections"
}

func GetCollections(db *gorm.DB) ([]Collection, error) {
	var colls []Collection
	err := db.Table("collections").Find(&colls).Error
	return colls, err
}

func SaveCollection(db *gorm.DB, coll *Collection) error {
	return db.Table(coll.TableName()).Save(coll).Error
}

// DeleteCollection soft deletes by primary key, so rows stay recoverable and
// GetCollections stops returning them.
func DeleteCollection(db *gorm.DB, id uint) error {
	return db.Delete(&Collection{}, id).Error
}
