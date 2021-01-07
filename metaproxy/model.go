package metaproxy

import (
	"github.com/jinzhu/gorm"
)
type User struct {
	gorm.Model
	UserId string `gorm:"primary_key"`
	UserName string
	UserToken string
}

type Bucket struct {
	gorm.Model
	BucketId string `gorm:"primary_key"`
	BucketName string
	UserId string
	Reader string
	Writer string
}

type Object struct {
	gorm.Model
	ObjectId string `gorm:"primary_key"`
	ObjectName string
	BucketId string
	UserId string
	FileName string
	Location string
	IsDelete bool
}