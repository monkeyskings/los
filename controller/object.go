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
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "object upload permission deny", w)
		utils.Logger.Info("req:" + requestid + ", object upload permission deny")
		return
	}
	bucketname := req.PostFormValue("bucketname")
	bucketid := utils.MakeStringMd5(bucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	objectname := req.PostFormValue("objectname")
	if utils.CheckFileNameNormal(objectname) == false{
		SendReponseMsg(requestid, http.StatusBadRequest, "object name abnormal", w)
		utils.Logger.Info("req:" + requestid + ", object name abnormal")
		return 
	}
	if BucketIsExist(bucketname, userid) == false {
		SendReponseMsg(requestid, http.StatusBadRequest, "object upload bucket not exists", w)
		utils.Logger.Info("req:" + requestid + ", object upload bucket not exits")
		return 
	}

	if ObjectIsExist(objectname, bucketid, userid){
		SendReponseMsg(requestid, http.StatusBadRequest, "object already exists", w)
		utils.Logger.Info("req:" + requestid + ", object already exits : ", username, bucketname, objectname)
		return 
	}
	file, _, err :=req.FormFile("filepath")
	nowtime := time.Now().Unix()
	filename := utils.MakeStringMd5(fmt.Sprintf("%s-%s-%s-%d", username, bucketname, objectname, nowtime))
	err = dataproxy.DataCreate(filename, file)
	if err != nil {
		SendReponseMsg(requestid, http.StatusInternalServerError, "object upload error", w)
		utils.Logger.Error("req:" + requestid + ", object upload error : ", username, bucketname, objectname, err)
		return 
	}
	objectid := utils.MakeStringMd5(objectname)
	espaddr, ok := GlobalConf["espaddr"]
	if ok == false{
		SendReponseMsg(requestid, http.StatusInternalServerError, "object upload error", w)
		utils.Logger.Error("req:" + requestid + ", object upload error : ", username, bucketname, objectname, err)
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
		SendReponseMsg(requestid, http.StatusInternalServerError, "object upload error", w)
		utils.Logger.Error("req:" + requestid + ",object upload error : ", username, bucketname, objectname, err)
		return 
	}
	SendReponseMsg(requestid, http.StatusOK, "object upload success", w)
	utils.Logger.Info("req:" + requestid + ", object upload successs : ", username, bucketname, objectname)
}

func ObjectProxyDownload(location string, w http.ResponseWriter, req *http.Request, requestid string, username string, bucketname string, objectname string)error{
	client := &http.Client{}

	body := fmt.Sprintf("{\"bucketname\":\"%s\",\"objectname\":\"%s\"}", bucketname, objectname)
   	
   	fmt.Printf("%s\n",body)
	downloadurl := fmt.Sprintf("http://%s/object/download", location)
	newreq, err := http.NewRequest("GET", downloadurl, strings.NewReader(body))
	if err != nil {
		utils.Logger.Error("req:" + requestid + ", object proxy new request error : ", username, bucketname, objectname, err)
		return err
	}
	for k, v := range req.Header {
        newreq.Header.Set(k, v[0])
    }
	res, err := client.Do(newreq)
	if err != nil {
		utils.Logger.Error("req:" + requestid + ", object proxy httpclient handle error : ", username, bucketname, objectname, err)
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
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "object download permission deny", w)
		utils.Logger.Info("req:" + requestid + ", object download permission deny")
		return 
	}
	var args ObjectDownloadArgs
	err := ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(requestid, http.StatusBadRequest, "object download args error", w)
		utils.Logger.Info("req:" + requestid + ", object download args error")
		return 
	}
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	objectname := args.ObjectName
	objectid := utils.MakeStringMd5(objectname)
	if ObjectIsExist(objectname, bucketid, userid) == false{
		SendReponseMsg(requestid, http.StatusBadRequest, "object is not exists", w)
		utils.Logger.Info("req:" + requestid + ", object is not exists")
		return 
	}
	var object metaproxy.Object
	Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", objectid, bucketid, userid).First(&object)
	location := object.Location
	filename := object.FileName
	espaddr, ok := GlobalConf["espaddr"]
	if ok == false{
		SendReponseMsg(requestid, http.StatusInternalServerError, "object download read error", w)
		utils.Logger.Error("req:" + requestid + ", object download read error : ", username, bucketname, objectname, err)
		return
	}
	if espaddr != location {
		err := ObjectProxyDownload(location, w, req, requestid, username, bucketname, objectname)
		if err != nil {
			SendReponseMsg(requestid, http.StatusInternalServerError, "object download read error", w)
			utils.Logger.Error("req:" + requestid + ", object download read error : ", username, bucketname, objectname, err)
			return
		}
		return 
	}
	err = dataproxy.DataRead(filename, w)
	if err != nil {
		SendReponseMsg(requestid, http.StatusInternalServerError, "object download read error", w)
		utils.Logger.Error("req:" + requestid + ", object download read error : ", username, bucketname, objectname, err)
		return
	}
	utils.Logger.Info("req:" + requestid + ", object download success : ", username, bucketname, objectname)
}

func ObjectList(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "object list permission deny", w)
		utils.Logger.Info("req:" + requestid + ", object list permission deny")
		return 
	}
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	var args ObjectListArgs
	err := ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(requestid, http.StatusBadRequest, "object list args error", w)
		utils.Logger.Info("req:" + requestid + ", object list args error")
		return 
	}
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	if BucketIsExist(bucketname, userid) == false {
		SendReponseMsg(requestid, http.StatusBadRequest, "object list bucket not exists", w)
		utils.Logger.Info("req:" + requestid + ", object list bucket not exists")
		return 
	}
	var objects []metaproxy.Object
	Dbcon.Where("user_id = ? and bucket_id = ? and is_delete = 0", userid, bucketid).Select("object_name").Find(&objects)
	msg := "objects :|"
	for _, object := range objects{
		msg = fmt.Sprintf("%s%s|", msg, object.ObjectName)
	}
	if len(objects) == 0 {
		msg = "no objects"
	}
    SendReponseMsg(requestid, http.StatusOK, msg, w)
    utils.Logger.Info("req:" + requestid + ", object list success : ", username, bucketname)
}

func ObjectDelete(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "object delete permission deny", w)
		utils.Logger.Info("req:" + requestid + ", object delete permission deny")
		return
	}
	var args ObjectDeleteArgs
	err := ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(requestid, http.StatusBadRequest, "object delete args error", w)
		utils.Logger.Info("req:" + requestid + ", object delete args error")
		return 
	}
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	objectname := args.ObjectName
	objectid := utils.MakeStringMd5(objectname)
	if ObjectIsExist(objectname, bucketid, userid) == false{
		SendReponseMsg(requestid, http.StatusBadRequest, "object is not exists", w)
		utils.Logger.Info("req:" + requestid + ", object is not exits")
		return 
	}

	err = Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", objectid, bucketid, userid).Update("is_delete", 1).Error
	if err != nil{
		SendReponseMsg(requestid, http.StatusInternalServerError, "object delete failed", w)
		utils.Logger.Error("req:" + requestid + ", object delete failed : ", username, bucketname, objectname, err)
		return 
	}
	SendReponseMsg(requestid, http.StatusOK, "object delete success", w)
	utils.Logger.Info("req:" + requestid + ", object delete success : ", username, bucketname, objectname)
}

func ObjectRename(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "object rename permission deny", w)
		utils.Logger.Info("req:" + requestid + ", object rename permission deny")
		return
	}
	var args ObjectRenameArgs
	err := ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(requestid, http.StatusBadRequest, "object rename args error", w)
		utils.Logger.Info("req:" + requestid + ", object rename args error")
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
		SendReponseMsg(requestid, http.StatusBadRequest, "destobject name abnormal", w)
		utils.Logger.Info("req:" + requestid + ", destobject name abnormal")
		return 
	}
	if ObjectIsExist(srcobjectname, bucketid, userid) == false{
		SendReponseMsg(requestid, http.StatusBadRequest, "object is not exists", w)
		utils.Logger.Info("req:" + requestid + ", object is not exists")
		return 
	}
	err = Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", srcobjectid, bucketid, userid).Updates(metaproxy.Object{ObjectName: destobjectname, ObjectId: destobjectid}).Error
	if err != nil{
		SendReponseMsg(requestid, http.StatusInternalServerError, "object rename failed", w)
		utils.Logger.Error("req:" + requestid + ", object rename failed : ", username, bucketname, srcobjectname, destobjectname, err)
		return 
	}
	SendReponseMsg(requestid, http.StatusOK, "object rename success", w)
	utils.Logger.Info("req:" + requestid + ", object rename success : ", username, bucketname, srcobjectname, destobjectname)
}

func ObjectMove(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "object move permission deny", w)
		utils.Logger.Info("req:" + requestid + ", object move permission deny")
		return
	}
	var args ObjectMoveArgs
	err := ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(requestid, http.StatusBadRequest, "object move args error", w)
		utils.Logger.Info("req:" + requestid + ",object move args error")
		return 
	}
	srcbucketname := args.SrcBucketName
	srcbucketid := utils.MakeStringMd5(srcbucketname)
	destbucketname := args.DestBucketName
	destbucketid := utils.MakeStringMd5(destbucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	if BucketIsExist(destbucketname, userid) == false {
		SendReponseMsg(requestid, http.StatusBadRequest, "destbucket not exists", w)
		utils.Logger.Info("req:" + requestid + ", destbucket not exists")
		return 
	}
	objectname := args.ObjectName
	objectid := utils.MakeStringMd5(objectname)
	if ObjectIsExist(objectname, srcbucketid, userid) == false{
		SendReponseMsg(requestid, http.StatusBadRequest, "object is not exists", w)
		utils.Logger.Info("req:" + requestid + ", object is not exists")
		return 
	}
	err = Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", objectid, srcbucketid, userid).Update("bucket_id", destbucketid).Error
	if err != nil{
		SendReponseMsg(requestid, http.StatusInternalServerError, "object move failed", w)
		utils.Logger.Error("req:" + requestid + ", object move failed : ", username, srcbucketname, destbucketname, objectname, err)
		return 
	}
	SendReponseMsg(requestid, http.StatusOK, "object move success", w)
	utils.Logger.Info("req:" + requestid + ", object move success : ", username, srcbucketname, destbucketname, objectname)
}
