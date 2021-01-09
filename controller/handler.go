package controller


import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/jinzhu/gorm"
	"los/utils"
	"errors"
	"io"
	"encoding/json"
	"fmt"
)

type ResponseData struct{
	Errorcode int
	Message string
}

type UserCreateArgs struct{
	Username string `json:"username"`
}

type BucketCreateArgs struct{
	BucketName string `json:"bucketname"`
}

type BucketDeleteArgs struct{
	BucketName string `json:"bucketname"`
}

type BucketRenameArgs struct{
	SrcBucketName string `json:"srcbucketname"`
	DestBucketName string `json:"destbucketname"`
}

type ObjectDownloadArgs struct {
	BucketName string `json:"bucketname"`
	ObjectName string `json:"objectname"`
}

type ObjectListArgs struct {
	BucketName string `json:"bucketname"`
}

type ObjectDeleteArgs struct {
	BucketName string `json:"bucketname"`
	ObjectName string `json:"objectname"`
}

type ObjectRenameArgs struct {
	BucketName string `json:"bucketname"`
	SrcObjectName string `json:"srcobjectname"`
	DestObjectName string `json:"destobjectname"`
}

type ObjectMoveArgs struct {
	SrcBucketName string `json:"srcbucketname"`
	DestBucketName string `json:"destbucketname"`
	ObjectName string `json:"objectname"`
}

var Dbcon *gorm.DB
var GlobalConf map[string] string

func RegisterHandlers() *httprouter.Router{
	router := httprouter.New()
	//user router
	router.POST("/user/create", UserCreate)
	router.PUT("/user/updatetoken", UserUpdateToken)

	//bucket router
	router.POST("/bucket/create", BucketCreate)
	router.GET("/bucket/list", BucketList)
	router.DELETE("/bucket/delete", BucketDelete)
	router.PUT("/bucket/rename", BucketRename)

	//object router
	router.POST("/object/upload", ObjectUpload)
	router.GET("/object/download", ObjectDownload)
	router.GET("/object/list", ObjectList)
	router.DELETE("/object/delete", ObjectDelete)
	router.PUT("/object/rename", ObjectRename)
	router.PUT("/object/move", ObjectMove)
	return router
}

func SendReponseMsg(errcode int, msg string, w http.ResponseWriter) {
	res := ResponseData{
		Errorcode: errcode,
		Message: msg,
	}
	ret, _ := json.Marshal(res)
	io.WriteString(w, string(ret))
}

func Start(Db *gorm.DB, gconf map[string]string)error{
	Dbcon = Db 
	GlobalConf = gconf
	listenport, ok := gconf["listenport"]
	localaddr, err := utils.GetLocalIpaddr()
	espaddr := fmt.Sprintf("%s:%s", localaddr, listenport)
	GlobalConf["espaddr"] = espaddr
	if err != nil{
		return err
	}
	if ok == false{
		return errors.New("global conf no listenport")
	}
	r := RegisterHandlers()
	utils.Logger.Info("controller start and listen request, server addr is : ", espaddr)
	return http.ListenAndServe(fmt.Sprintf(":%s", listenport), r)
}
