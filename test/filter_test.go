package gormfilter

import (
	"testing"

	gormfilter "github.com/beratyavuzyigit/gorm-filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Author struct {
	ID   int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code string `gorm:"column:code" json:"code"`
	Name string `gorm:"column:name" json:"name"`
}

type Book struct {
	ID         int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code       string `gorm:"column:code" json:"code"`
	Name       string `gorm:"column:name" json:"name"`
	FkAuthorId int    `gorm:"column:fk_author_id" json:"fk_author_id"`
	Author     Author `gorm:"foreignKey:FkAuthorId"`
}

func newDB() *gormfilter.DB {
	//sqlite
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	return gormfilter.NewDB(db)
}

func Test_Filter(t *testing.T) {
	db := newDB()
	db.AutoMigrate(&Author{}, &Book{})
	db.Create(&Author{ID: 1, Code: "A1", Name: "Author 1"})
	db.Create(&Author{ID: 2, Code: "A2", Name: "Author 2"})
	db.Create(&Book{ID: 1, Code: "B1", Name: "Book 1", FkAuthorId: 1})
	db.Create(&Book{ID: 2, Code: "B2", Name: "Book 2", FkAuthorId: 1})
	db.Create(&Book{ID: 3, Code: "B3", Name: "Book 3", FkAuthorId: 2})

	t.Run("Test_Filter", func(t *testing.T) {
		var model []Book
		result := db.Model(Book{}).Query(map[string]string{"fk_author_id__code": "A1"}).Find(&model)
		if result.Error != nil {
			t.Error(result.Error)
		}
		if len(model) != 2 {
			t.Errorf("Expected 2, got %d", len(model))
		}
	})
}
