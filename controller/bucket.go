package controller

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"los/utils"
	"los/metaproxy"
	"fmt"
	"strconv"
	"time"
)

func BucketIsExist(bucketname, userid string) bool{
	var count int
	Dbcon.Model(&metaproxy.Bucket{}).Where("bucket_name = ? and user_id = ?", bucketname, userid).Count(&count)
	if count > 0{
		return true
	}
	return false
}

func BucketHasObjects(bucketid string) bool{
	var count int
	Dbcon.Model(&metaproxy.Object{}).Where("bucket_id = ? and is_delete = 0", bucketid).Count(&count)
	if count > 0{
		return true
	}
	return false
}

func BucketNumLimit(userid string) bool{
	var count int
	bucketnumconf, ok := GlobalConf["bucketnum"]
	if ok == false {
		utils.Logger.Error("global conf bucketnum get error")
		return false
	}
	bucketnum, err := strconv.Atoi(bucketnumconf)
	if err != nil {
		utils.Logger.Error("global conf bucketnum trans int error")
		return false
	}
	Dbcon.Model(&metaproxy.Bucket{}).Where("user_id = ? ", userid).Count(&count)
	if count < bucketnum{
		return false
	}
	return true
}

func BucketCreate(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "bucket create permission deny", w)
		utils.Logger.Info("req:" + requestid + ", bucket create permission deny")
		return
	}
	var args BucketCreateArgs
	err := ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(requestid, http.StatusBadRequest, "bucket create args error", w)
		utils.Logger.Info("req:" + requestid + ",bucket create args error")
		return 
	}
	bucketname := args.BucketName
	if utils.CheckNameNormal(bucketname) == false{
		SendReponseMsg(requestid, http.StatusBadRequest, "bucket name abnormal", w)
		utils.Logger.Info("req:" + requestid + ", bucket name abnormal")
		return 
	}
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	if BucketIsExist(bucketname, userid) {
		SendReponseMsg(requestid, http.StatusBadRequest, "bucket aready exists", w)
		utils.Logger.Info("req:" + requestid + ", bucket aready exists : ", username, bucketname)
		return 
	}
	if BucketNumLimit(userid){
		SendReponseMsg(requestid, http.StatusBadRequest, "bucket limit exceed", w)
		utils.Logger.Info("req:" + requestid + ", bucket limit exceed : ", username, bucketname)
		return 
	}
	bucketid := utils.MakeStringMd5(bucketname)
	bucket := metaproxy.Bucket{
		BucketId : bucketid,
		BucketName : bucketname,
		UserId : userid,
	}
	err = Dbcon.Create(&bucket).Error

	if err != nil{
		SendReponseMsg(requestid, http.StatusInternalServerError, "bucket create failed", w)
		utils.Logger.Error("req:" + requestid + ", bucket create failed : ", username, bucketname, err)
		return 
	}
	SendReponseMsg(requestid, http.StatusOK, "bucket create success", w)
	utils.Logger.Info("req:" + requestid + ", bucket create success : ", username, bucketname)
}


func BucketList(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "bucket list permission deny", w)
		utils.Logger.Info("req:" + requestid + ", bucket list permission deny")
		return 
	}
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	var bucekts []metaproxy.Bucket
	Dbcon.Where("user_id = ?", userid).Select("bucket_name").Find(&bucekts)
	msg := "bucket :|"
	for _, bucket := range bucekts{
		msg = fmt.Sprintf("%s%s|", msg, bucket.BucketName)
	}
	if len(bucekts) == 0{
		msg = "no buckets"
	}
	SendReponseMsg(requestid, http.StatusOK, msg, w)
    utils.Logger.Info("req:" + requestid + ", bucket list success : ", username)
}

func BucketDelete(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "bucket delete permission deny", w)
		utils.Logger.Info("req:" + requestid + ", bucket delete permission deny")
		return 
	}
	var args BucketDeleteArgs
	err := ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(requestid, http.StatusBadRequest, "bucket delete args error", w)
		utils.Logger.Info("req:" + requestid + ", bucket delete args error")
		return 
	}
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	if BucketIsExist(bucketname, userid) == false{
		SendReponseMsg(requestid, http.StatusBadRequest, "bucket is not exists", w)
		utils.Logger.Info("req:" + requestid + ", bucket is not exists")
		return
	}
	if BucketHasObjects(bucketid) {
		SendReponseMsg(requestid, http.StatusBadRequest, "bucket has some objects", w)
		utils.Logger.Info("req:" + requestid + ", bucket has some objects")
		return
	}
	err = Dbcon.Where("bucket_id = ? and user_id = ?", bucketid, userid).Delete(metaproxy.Bucket{}).Error
	if err != nil{
		SendReponseMsg(requestid, http.StatusInternalServerError, "bucket delete failed", w)
		utils.Logger.Error("req:" + requestid + ", bucket delete failed : ", username, bucketname, err)
		return 
	}
    SendReponseMsg(requestid, http.StatusOK, "bucket delete success", w)
    utils.Logger.Info("req:" + requestid + ", bucket delete success : ", username, bucketname)
}

func BucketRename(w http.ResponseWriter, req *http.Request, p httprouter.Params){
	requestid := utils.MakeStringMd5(fmt.Sprintf("%d", time.Now().Unix()))
	if UserConfirm(req.Header) != true{
		SendReponseMsg(requestid, http.StatusUnauthorized, "bucket rename permission deny", w)
		utils.Logger.Info("req:" + requestid + ", bucket rename permission deny")
		return 
	}
	var args BucketRenameArgs
	err := ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(requestid, http.StatusBadRequest, "bucket rename args error", w)
		utils.Logger.Info("req:" + requestid + ", bucket rename args error")
		return 
	}
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	srcbucketname := args.SrcBucketName
	srcbucketid := utils.MakeStringMd5(srcbucketname)
	destbucketname := args.DestBucketName
	destbucketid := utils.MakeStringMd5(destbucketname)
	if utils.CheckNameNormal(destbucketname) == false{
		SendReponseMsg(requestid, http.StatusBadRequest, "destbucket name abnormal", w)
		utils.Logger.Info("req:" + requestid + ", destbucket name abnormal")
		return 
	}
	if BucketIsExist(srcbucketname, userid) == false{
		SendReponseMsg(requestid, http.StatusBadRequest, "bucket is not exists", w)
		utils.Logger.Info("req:" + requestid + ", bucket is not exists")
		return
	}
	err = Dbcon.Model(&metaproxy.Bucket{}).Where("bucket_id = ? and user_id = ?", srcbucketid, userid).Updates(metaproxy.Bucket{BucketName: destbucketname, BucketId: destbucketid}).Error
	if err != nil {
		SendReponseMsg(requestid, http.StatusInternalServerError, "bucket rename failed", w)
		utils.Logger.Error("req:" + requestid + ",bucket rename failed : ", username, srcbucketname, destbucketname, err)
		return 
	}
	err = Dbcon.Model(&metaproxy.Object{}).Where("bucket_id = ? and user_id = ?", srcbucketid, userid).Update("bucket_id", destbucketid).Error
	if err != nil{
		SendReponseMsg(requestid, http.StatusInternalServerError, "bucket rename failed", w)
		utils.Logger.Error("req:" + requestid + ", bucket rename failed : ", username, srcbucketname, destbucketname, err)
		return 
	}
    SendReponseMsg(requestid, http.StatusOK, "bucket rename success", w)
    utils.Logger.Info("req:" + requestid + ", bucket rename success : ", username, srcbucketname, destbucketname)
}