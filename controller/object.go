package controller


import (
	"net/http"
	"io"
	"github.com/julienschmidt/httprouter"
	"los/utils"
	"encoding/json"
	"fmt"
	"strings"
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
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "object upload permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object upload permission deny")
		return
	}
	bucketname := req.PostFormValue("bucketname")
	bucketid := utils.MakeStringMd5(bucketname)
	username := req.Header["Username"][0]
	userid := utils.MakeStringMd5(username)
	objectname := req.PostFormValue("objectname")
	if utils.CheckFileNameNormal(objectname) == false{
		res.Errorcode = 400
		res.Message = "object name abnormal"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object name abnormal")
		return 
	}
	if BucketIsExist(bucketname, userid) == false {
		res.Errorcode = 400
		res.Message = "object upload bucket not exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object upload bucket not exits")
		return 
	}

	if ObjectIsExist(objectname, bucketid, userid){
		res.Errorcode = 400
		res.Message = "object already exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object already exits : ", username, bucketname, objectname)
		return 
	}
	file, _, err :=req.FormFile("filepath")
	filename := utils.MakeStringMd5(fmt.Sprintf("%s-%s-%s", username, bucketname, objectname))
	err = dataproxy.DataCreate(filename, file)
	if err != nil {
		res.Errorcode = 500
		res.Message = "object upload error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Error("object upload error : ", username, bucketname, objectname, err)
		return 
	}
	objectid := utils.MakeStringMd5(objectname)
	espaddr, ok := GlobalConf["espaddr"]
	if ok == false{
		res.Errorcode = 500
		res.Message = "object upload error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
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
	Dbcon.Create(&object)
	if ObjectIsExist(objectname, bucketid, userid) == false{
		res.Errorcode = 500
		res.Message = "object upload error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Error("object upload error : ", username, bucketname, objectname, err)
		return 
	}
	res.Message = "object upload success"
	ret, _ := json.Marshal(res)
	io.WriteString(w, string(ret))
}

func ObjectProxyDownload(location string, w http.ResponseWriter, req *http.Request, bucketname string, objectname string)error{
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
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "object download permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object download permission deny")
		return 
	}
	var args ObjectDownloadArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		res.Errorcode = 400
		res.Message = "object download args error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
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
		res.Errorcode = 400
		res.Message = "object is not exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object is not exists")
		return 
	}
	var object metaproxy.Object
	Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", objectid, bucketid, userid).First(&object)
	location := object.Location
	filename := object.FileName
	espaddr, ok := GlobalConf["espaddr"]
	fmt.Println(filename, location)
	if ok == false{
		res.Errorcode = 500
		res.Message = "object download read error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Error("object download read error : ", username, bucketname, objectname, err)
		return
	}
	if espaddr != location {
		err := ObjectProxyDownload(location, w, req, bucketname, objectname)
		if err != nil {
			res.Errorcode = 500
			res.Message = "object download read error"
			ret, _ := json.Marshal(res)
			io.WriteString(w, string(ret))
			utils.Logger.Error("object download read error : ", username, bucketname, objectname, err)
			return
		}
		return 
	}
	err = dataproxy.DataRead(filename, w)
	if err != nil {
		res.Errorcode = 500
		res.Message = "object download read error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Error("object download read error : ", username, bucketname, objectname, err)
		return
	}
}

func ObjectList(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "object list permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object list permission deny")
		return 
	}
	userid := utils.MakeStringMd5(req.Header["Username"][0])
	var args ObjectListArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		res.Errorcode = 400
		res.Message = "object list args error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object list args error")
		return 
	}
	bucketname := args.BucketName
	bucketid := utils.MakeStringMd5(bucketname)
	if BucketIsExist(bucketname, userid) == false {
		res.Errorcode = 400
		res.Message = "object list bucket not exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object list bucket not exists")
		return 
	}
	var objects []metaproxy.Object
	Dbcon.Where("user_id = ? and bucket_id = ? and is_delete = 0", userid, bucketid).Select("object_name").Find(&objects)
	msg := "objects :"
	for _, object := range objects{
		msg = fmt.Sprintf("%s %s", msg, object.ObjectName)
	}
	res.Message = msg
	ret, _ := json.Marshal(res)
    io.WriteString(w, string(ret))
}

func ObjectDelete(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "object delete permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object delete permission deny")
		return
	}
	var args ObjectDeleteArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		res.Errorcode = 400
		res.Message = "object delete args error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
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
		res.Errorcode = 400
		res.Message = "object is not exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object is not exits")
		return 
	}

	Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", objectid, bucketid, userid).Update("is_delete", 1)
	res.Message = "object delete success"
	ret, _ := json.Marshal(res)
	io.WriteString(w, string(ret))
}

func ObjectRename(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "object rename permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object rename permission deny")
		return
	}
	var args ObjectRenameArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		res.Errorcode = 400
		res.Message = "object rename args error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
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
		res.Errorcode = 400
		res.Message = "destobject name abnormal"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("destobject name abnormal")
		return 
	}
	if ObjectIsExist(srcobjectname, bucketid, userid) == false{
		res.Errorcode = 400
		res.Message = "object is not exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object is not exists")
		return 
	}
	Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", srcobjectid, bucketid, userid).Updates(metaproxy.Object{ObjectName: destobjectname, ObjectId: destobjectid})
	res.Message = "object rename success"
	ret, _ := json.Marshal(res)
	io.WriteString(w, string(ret))
}

func ObjectMove(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "object move permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object move permission deny")
		return
	}
	var args ObjectMoveArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		res.Errorcode = 400
		res.Message = "object move args error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
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
		res.Errorcode = 400
		res.Message = "destbucket not exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("destbucket not exists")
		return 
	}
	objectname := args.ObjectName
	objectid := utils.MakeStringMd5(objectname)
	if ObjectIsExist(objectname, srcbucketid, userid) == false{
		res.Errorcode = 400
		res.Message = "object is not exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("object is not exists")
		return 
	}
	Dbcon.Model(&metaproxy.Object{}).Where("object_id = ? and bucket_id = ? and user_id = ?", objectid, srcbucketid, userid).Update("bucket_id", destbucketid)
	res.Message = "object rename success"
	ret, _ := json.Marshal(res)
	io.WriteString(w, string(ret))
}
