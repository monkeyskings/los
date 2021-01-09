package controller


import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"los/utils"
	"fmt"
	"time"
	"strings"
	"io"
	"los/metaproxy"
	"los/dataproxy"
)

func ObjectIsExist(objectname string, bucketid string, userid string) bool{
	var count int
	Dbcon.Model(&metaproxy.Object{}).Where("object_name = ? and bucket_id = ? and user_id = ?", objectname, bucketid, userid).Count(&count)
	if count > 0{
		return true
	}
	return false
}

func ObjectUpload(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if UserConfirm(req.Header) != true{
		SendReponseMsg(http.StatusUnauthorized, "object upload permission deny", w)
		utils.Logger.Info("object upload permission deny")
		return
	}
	bucketname := req.PostFormValue("bucketname")
	bucketid := utils.MakeStringMd5(bucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	objectname := req.PostFormValue("objectname")
	if utils.CheckFileNameNormal(objectname) == false{
		SendReponseMsg(http.StatusBadRequest, "object name abnormal", w)
		utils.Logger.Info("object name abnormal")
		return 
	}
	if BucketIsExist(bucketname, userid) == false {
		SendReponseMsg(http.StatusBadRequest, "object upload bucket not exists", w)
		utils.Logger.Info("object upload bucket not exits")
		return 
	}

	if ObjectIsExist(objectname, bucketid, userid){
		SendReponseMsg(http.StatusBadRequest, "object already exists", w)
		utils.Logger.Info("object already exits : ", username, bucketname, objectname)
		return 
	}
	file, _, err :=req.FormFile("filepath")
	nowtime := time.Now().Unix()
	filename := utils.MakeStringMd5(fmt.Sprintf("%s-%s-%s-%d", username, bucketname, objectname, nowtime))
	err = dataproxy.DataCreate(filename, file)
	if err != nil {
		SendReponseMsg(http.StatusInternalServerError, "object upload error", w)
		utils.Logger.Error("object upload error : ", username, bucketname, objectname, err)
		return 
	}
	objectid := utils.MakeStringMd5(objectname)
	espaddr, ok := GlobalConf["espaddr"]
	if ok == false{
		SendReponseMsg(http.StatusInternalServerError, "object upload error", w)
		utils.Logger.Error("object upload error : ", username, bucketname, objectname, err)
		return 
	}
	object := metaproxy.Object{
		ObjectId : objectid,
		ObjectName : objectname,
		BucketId : bucketid,
		UserId : userid,
		FileName : filename,
		IsDelete : false,
		Location : espaddr,
	}
	err = Dbcon.Create(&object).Error
	if err != nil{
		SendReponseMsg(http.StatusInternalServerError, "object upload error", w)
		utils.Logger.Error("object upload error : ", username, bucketname, objectname, err)
		return 
	}
	SendReponseMsg(http.StatusOK, "object upload success", w)
}

func ObjectProxyDownload(location string, w http.ResponseWriter, req *http.Request, username string, bucketname string, objectname string)error{
	client := &http.Client{}

	body := fmt.Sprintf("{\"bucketname\":\"%s\",\"objectname\":\"%s\"}", bucketname, objectname)
   	
   	fmt.Printf("%s\n",body)
	downloadurl := fmt.Sprintf("http://%s/object/download", location)
	newreq, err := http.NewRequest("GET", downloadurl, strings.NewReader(body))
	if err != nil {
		utils.Logger.Error("object proxy new request error : ", username, bucketname, objectname, err)
		return err
	}
	for k, v := range req.Header {
        newreq.Header.Set(k, v[0])
    }
	res, err := client.Do(newreq)
	if err != nil {
		utils.Logger.Error("object proxy httpclient handle error : ", username, bucketname, objectname, err)
		return err
	}
	defer res.Body.Close()
	for k, v := range res.Header {
        w.Header().Set(k, v[0])
    }
	io.Copy(w, res.Body)
	return nil
}

func ObjectDownload(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if UserConfirm(req.Header) != true{
		SendReponseMsg(http.StatusUnauthorized, "object download permission deny", w)
		utils.Logger.Info("object download permission deny")
		return 
	}
	var args ObjectDownloadArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(http.StatusBadRequest, "object download args error", w)
		utils.Logger.Info("object download args error")
		return 
	}
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	objectname := args.ObjectName
	objectid := utils.MakeStringMd5(objectname)
	fmt.Println(username, bucketname, objectname)
	if ObjectIsExist(objectname, bucketid, userid) == false{
		SendReponseMsg(http.StatusBadRequest, "object is not exists", w)
		utils.Logger.Info("object is not exists")
		return 
	}
	var object metaproxy.Object
	Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", objectid, bucketid, userid).First(&object)
	location := object.Location
	filename := object.FileName
	espaddr, ok := GlobalConf["espaddr"]
	if ok == false{
		SendReponseMsg(http.StatusInternalServerError, "object download read error", w)
		utils.Logger.Error("object download read error : ", username, bucketname, objectname, err)
		return
	}
	if espaddr != location {
		err := ObjectProxyDownload(location, w, req, username, bucketname, objectname)
		if err != nil {
			SendReponseMsg(http.StatusInternalServerError, "object download read error", w)
			utils.Logger.Error("object download read error : ", username, bucketname, objectname, err)
			return
		}
		return 
	}
	err = dataproxy.DataRead(filename, w)
	if err != nil {
		SendReponseMsg(http.StatusInternalServerError, "object download read error", w)
		utils.Logger.Error("object download read error : ", username, bucketname, objectname, err)
		return
	}
}

func ObjectList(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if UserConfirm(req.Header) != true{
		SendReponseMsg(http.StatusUnauthorized, "object list permission deny", w)
		utils.Logger.Info("object list permission deny")
		return 
	}
	userid := utils.MakeStringMd5(req.Header["Username"][0])
	var args ObjectListArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(http.StatusBadRequest, "object list args error", w)
		utils.Logger.Info("object list args error")
		return 
	}
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	if BucketIsExist(bucketname, userid) == false {
		SendReponseMsg(http.StatusBadRequest, "object list bucket not exists", w)
		utils.Logger.Info("object list bucket not exists")
		return 
	}
	var objects []metaproxy.Object
	Dbcon.Where("user_id = ? and bucket_id = ? and is_delete = 0", userid, bucketid).Select("object_name").Find(&objects)
	msg := "objects :"
	for _, object := range objects{
		msg = fmt.Sprintf("%s %s", msg, object.ObjectName)
	}
    SendReponseMsg(http.StatusOK, msg, w)
}

func ObjectDelete(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if UserConfirm(req.Header) != true{
		SendReponseMsg(http.StatusUnauthorized, "object delete permission deny", w)
		utils.Logger.Info("object delete permission deny")
		return
	}
	var args ObjectDeleteArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(http.StatusBadRequest, "object delete args error", w)
		utils.Logger.Info("object delete args error")
		return 
	}
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	objectname := args.ObjectName
	objectid := utils.MakeStringMd5(objectname)
	if ObjectIsExist(objectname, bucketid, userid) == false{
		SendReponseMsg(http.StatusBadRequest, "object is not exists", w)
		utils.Logger.Info("object is not exits")
		return 
	}

	err = Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", objectid, bucketid, userid).Update("is_delete", 1).Error
	if err != nil{
		SendReponseMsg(http.StatusInternalServerError, "object delete failed", w)
		utils.Logger.Error("object delete failed : ", username, bucketname, objectname, err)
		return 
	}
	SendReponseMsg(http.StatusOK, "object delete success", w)
}

func ObjectRename(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if UserConfirm(req.Header) != true{
		SendReponseMsg(http.StatusUnauthorized, "object rename permission deny", w)
		utils.Logger.Info("object rename permission deny")
		return
	}
	var args ObjectRenameArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(http.StatusBadRequest, "object rename args error", w)
		utils.Logger.Info("object rename args error")
		return 
	}
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	srcobjectname := args.SrcObjectName
	srcobjectid := utils.MakeStringMd5(srcobjectname)
	destobjectname := args.DestObjectName
	destobjectid := utils.MakeStringMd5(destobjectname)
	if utils.CheckFileNameNormal(destobjectname) == false{
		SendReponseMsg(http.StatusBadRequest, "destobject name abnormal", w)
		utils.Logger.Info("destobject name abnormal")
		return 
	}
	if ObjectIsExist(srcobjectname, bucketid, userid) == false{
		SendReponseMsg(http.StatusBadRequest, "object is not exists", w)
		utils.Logger.Info("object is not exists")
		return 
	}
	err = Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", srcobjectid, bucketid, userid).Updates(metaproxy.Object{ObjectName: destobjectname, ObjectId: destobjectid}).Error
	if err != nil{
		SendReponseMsg(http.StatusInternalServerError, "object rename failed", w)
		utils.Logger.Error("object rename failed : ", username, bucketname, srcobjectname, destobjectname, err)
		return 
	}
	SendReponseMsg(http.StatusOK, "object rename success", w)
}

func ObjectMove(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if UserConfirm(req.Header) != true{
		SendReponseMsg(http.StatusUnauthorized, "object move permission deny", w)
		utils.Logger.Info("object move permission deny")
		return
	}
	var args ObjectMoveArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(http.StatusBadRequest, "object move args error", w)
		utils.Logger.Info("object move args error")
		return 
	}
	srcbucketname := args.SrcBucketName
	srcbucketid := utils.MakeStringMd5(srcbucketname)
	destbucketname := args.DestBucketName
	destbucketid := utils.MakeStringMd5(destbucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	if BucketIsExist(destbucketname, userid) == false {
		SendReponseMsg(http.StatusBadRequest, "destbucket not exists", w)
		utils.Logger.Info("destbucket not exists")
		return 
	}
	objectname := args.ObjectName
	objectid := utils.MakeStringMd5(objectname)
	if ObjectIsExist(objectname, srcbucketid, userid) == false{
		SendReponseMsg(http.StatusBadRequest, "object is not exists", w)
		utils.Logger.Info("object is not exists")
		return 
	}
	err = Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", objectid, srcbucketid, userid).Update("bucket_id", destbucketid).Error
	if err != nil{
		SendReponseMsg(http.StatusInternalServerError, "object move failed", w)
		utils.Logger.Error("object move failed : ", username, srcbucketname, destbucketname, objectname, err)
		return 
	}
	SendReponseMsg(http.StatusOK, "object move success", w)
}
