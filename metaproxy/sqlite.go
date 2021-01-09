package metaproxy

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func Sqlite3Open(dbpath string) (*gorm.DB, error) {
	return gorm.Open("sqlite3", dbpath)
}