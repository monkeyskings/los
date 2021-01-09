package metaproxy

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)


func MysqlOpen(dbuser string, dbpass string, dbaddr string, dbname string)(*gorm.DB, error){
	connectstr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", dbuser, dbpass, dbaddr, dbname)
	return gorm.Open("mysql", connectstr)
}