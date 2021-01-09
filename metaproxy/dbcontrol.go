package metaproxy

import (
	"github.com/jinzhu/gorm"
	"errors"
	"los/utils"
)



func DbOpen(dbconf map[string]string) (*gorm.DB, error){
	dbtype, ok:= dbconf["dbtype"]
	if ok == false{
		return nil, errors.New("dbconf no dbtype")
	}
	if dbtype == "sqlite3"{
		dbpath, ok:= dbconf["dbname"]
		if ok == false{
			return nil, errors.New("sqlite dbconf no dbpath")
		}
		utils.Logger.Info("start open sqlite3")
		return Sqlite3Open(dbpath)
	}
	if dbtype == "mysql"{
		dbuser, ok1 := dbconf["dbuser"]
		dbpass, ok2 := dbconf["dbpass"]
		dbaddr, ok3 := dbconf["dbaddr"]
		dbname, ok4 := dbconf["dbname"]
		if (ok1 && ok2 && ok3 && ok4) == false{
			return  nil, errors.New("mysql dbconf error")
		}
		return MysqlOpen(dbuser, dbpass, dbaddr, dbname)
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

