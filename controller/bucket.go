package controller

import (
	"net/http"
	"io"
	"github.com/julienschmidt/httprouter"
	"los/utils"
	"encoding/json"
	"los/metaproxy"
	"fmt"
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

func BucketCreate(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "bucket create permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket create permission deny")
		return
	}
	var args BucketCreateArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		res.Errorcode = 400
		res.Message = "bucket create args error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket create args error")
		return 
	}
	bucketname := args.BucketName
	if utils.CheckNameNormal(bucketname) == false{
		res.Errorcode = 400
		res.Message = "bucket name abnormal"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket name abnormal")
		return 
	}
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	if BucketIsExist(bucketname, userid) {
		res.Errorcode = 400
		res.Message = "bucket aready exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket aready exists : ", username, bucketname)
		return 
	}
	bucketid := utils.MakeStringMd5(bucketname)
	bucket := metaproxy.Bucket{
		BucketId : bucketid,
		BucketName : bucketname,
		UserId : userid,
	}
	Dbcon.Create(&bucket)

	if BucketIsExist(bucketname, userid) != true{
		res.Errorcode = 500
		res.Message = "bucket create failed"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Error("bucket create failed : ", username, bucketname, err)
		return 
	}

	res.Message = "bucket create success"
	ret, _ := json.Marshal(res)
	io.WriteString(w, string(ret))
	utils.Logger.Info("bucket create success : ", username, bucketname)
}


func BucketList(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "bucket list permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket list permission deny")
		return 
	}
	userid := utils.MakeStringMd5(req.Header["Username"][0])
	var bucekts []metaproxy.Bucket
	Dbcon.Where("user_id = ?", userid).Select("bucket_name").Find(&bucekts)
	msg := "bucket :"
	for _, bucket := range bucekts{
		msg = fmt.Sprintf("%s %s", msg, bucket.BucketName)
	}
	res.Message = msg
	ret, _ := json.Marshal(res)
    io.WriteString(w, string(ret))
    utils.Logger.Info("bucket list success")
}

func BucketDelete(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "bucket delete permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket delete permission deny")
		return 
	}
	var args BucketDeleteArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		res.Errorcode = 400
		res.Message = "bucket delete args error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket delete args error")
		return 
	}
	userid := utils.MakeStringMd5(req.Header["Username"][0])
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	if BucketIsExist(bucketname, userid) == false{
		res.Errorcode = 400
		res.Message = "bucket is not exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket is not exists")
		return
	}
	if BucketHasObjects(bucketid) {
		res.Errorcode = 400
		res.Message = "bucket has some objects"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket has some objects")
		return
	}
	Dbcon.Where("bucket_id = ? and user_id = ?", bucketid, userid).Delete(metaproxy.Bucket{})
	res.Message = "bucket delete success"
	ret, _ := json.Marshal(res)
    io.WriteString(w, string(ret))
    utils.Logger.Info("bucket delete success")
}

func BucketRename(w http.ResponseWriter, req *http.Request, p httprouter.Params){
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "bucket rename permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket rename permission deny")
		return 
	}
	var args BucketRenameArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		res.Errorcode = 400
		res.Message = "bucket rename args error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket rename args error")
		return 
	}
	userid := utils.MakeStringMd5(req.Header["Username"][0])
	srcbucketname := args.SrcBucketName
	srcbucketid := utils.MakeStringMd5(srcbucketname)
	destbucketname := args.DestBucketName
	destbucektid := utils.MakeStringMd5(destbucketname)
	if utils.CheckNameNormal(destbucketname) == false{
		res.Errorcode = 400
		res.Message = "destbucket name abnormal"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("destbucket name abnormal")
		return 
	}
	if BucketIsExist(srcbucketname, userid) == false{
		res.Errorcode = 400
		res.Message = "bucket is not exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("bucket is not exists")
		return
	}
	Dbcon.Model(&metaproxy.Bucket{}).Where("bucket_id = ? and user_id = ?", srcbucketid, userid).Updates(metaproxy.Bucket{BucketName: destbucketname, BucketId: destbucektid})
	Dbcon.Model(&metaproxy.Object{}).Where("bucket_id = ? and user_id = ?", srcbucketid, userid).Update("bucket_id", destbucektid)
	res.Message = "bucket rename success"
	ret, _ := json.Marshal(res)
    io.WriteString(w, string(ret))
    utils.Logger.Info("bucket rename success")
}