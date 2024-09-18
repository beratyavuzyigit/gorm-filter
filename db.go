package gormfilter

import (
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(db *gorm.DB) *DB {
	return &DB{db}
}

func (db *DB) Model(model interface{}) *DB {
	return &DB{DB: db.DB.Model(model)}
}
