package metaproxy

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"errors"
	"los/utils"
)

func Sqlite3Open(dbpath string) (*gorm.DB, error) {
	return gorm.Open("sqlite3", dbpath)
}

func DbOpen(dbconf map[string]string) (*gorm.DB, error){
	dbtype, ok:= dbconf["dbtype"]
	if ok == false{
		return nil, errors.New("dbconf no dbtype")
	}
	if dbtype == "sqlite3"{
		dbpath, ok:= dbconf["dbname"]
		if ok == false{
			return nil, errors.New("dbconf no dbpath")
		}
		utils.Logger.Info("start open sqlite3")
		return Sqlite3Open(dbpath)
	}
	return nil, nil
}

func DbClose(db *gorm.DB) error{
	utils.Logger.Info("start close db")
	return db.Close()
}

func DbInitial(dbconf map[string]string) error{
	db, err := DbOpen(dbconf)
	if err != nil{
		return err
	}
	defer db.Close()
	utils.Logger.Info("start init db orm")
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Bucket{})
	db.AutoMigrate(&Object{})
	return nil
}

