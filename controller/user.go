package controller


import (
	"net/http"
	"time"
	"io"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"los/utils"
	"encoding/json"
	"math/rand"
	"los/metaproxy"
)

func UserIsExist(username string) bool{
	var count int
	Dbcon.Model(&metaproxy.User{}).Where("user_name = ?", username).Count(&count)
	if count > 0{
		return true
	}
	return false
}


func UserCreate(w http.ResponseWriter, req *http.Request, p httprouter.Params){
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	var args UserCreateArgs
	err := utils.ParseHttpBody(req.Body, &args)
	if err != nil{
		res.Errorcode = 400
		res.Message = "user create args error"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("user create args error")
		return 
	}
	username := args.Username
	
	if utils.CheckNameNormal(username) == false{
		res.Errorcode = 400
		res.Message = "user name abnormal"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("user name abnormal")
		return 
	}
	nowtime := time.Now().Unix()
	seedstr := fmt.Sprintf("%d-%s", nowtime, username)
	token := utils.MakeStringMd5(seedstr)
	userid := utils.MakeStringMd5(username)

	user := metaproxy.User{
		UserId: userid,
		UserName: username,
		UserToken: token,
	}
	
	if UserIsExist(username) {
		res.Errorcode = 400
		res.Message = "user aready exists"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("user aready exists : ", username)
		return 
	}
	Dbcon.Create(&user)
	if UserIsExist(username) != true{
		res.Errorcode = 500
		res.Message = "user create failed"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Error("user create failed : ", username)
		return 
	}
	res.Message = token
	ret, _ := json.Marshal(res)
	io.WriteString(w, string(ret))
	utils.Logger.Info("user create success : ", username)
}

func UserConfirm(header http.Header) bool{
	if len(header["Username"]) != 1 || len(header["Token"]) != 1{
		return false
	}
	username := header["Username"][0]
	token := header["Token"][0]
	var count int
	Dbcon.Model(&metaproxy.User{}).Where("user_name = ? and user_token = ?", username, token).Count(&count)
	if count > 0{
		return true
	}
	return false
}

func UserUpdateToken(w http.ResponseWriter, req *http.Request, p httprouter.Params){
	confirmret := UserConfirm(req.Header)
	res := ResponseData{
		Errorcode: 200,
		Message: "",
	}
	if confirmret != true{
		res.Errorcode = 400
		res.Message = "user update token permission deny"
		ret, _ := json.Marshal(res)
		io.WriteString(w, string(ret))
		utils.Logger.Info("user update token permission deny")
		return 
	}
	username := req.Header["Username"]
	token := utils.MakeStringMd5(fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Int()))
	Dbcon.Model(&metaproxy.User{}).Where("user_name = ?", username).Update("user_token", token)
	res.Message = token
	ret, _ := json.Marshal(res)
	io.WriteString(w, string(ret))
	utils.Logger.Info("user update token success")
}


