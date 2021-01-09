package controller


import (
	"net/http"
	"time"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"los/utils"
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
	var args UserCreateArgs
	err := ParseHttpBody(req.Body, &args)
	if err != nil{
		SendReponseMsg(http.StatusBadRequest, "user create args error", w)
		utils.Logger.Info("user create args error")
		return 
	}
	username := args.Username
	
	if utils.CheckNameNormal(username) == false{
		SendReponseMsg(http.StatusBadRequest, "user name abnormal", w)
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
		SendReponseMsg(http.StatusBadRequest, "user aready exists", w)
		utils.Logger.Info("user aready exists : ", username)
		return 
	}
	err = Dbcon.Create(&user).Error
	if err != nil{
		SendReponseMsg(http.StatusInternalServerError, "user create failed", w)
		utils.Logger.Error("user create failed : ", username)
		return 
	}
	SendReponseMsg(http.StatusOK, token, w)
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
	if UserConfirm(req.Header) != true{
		SendReponseMsg(http.StatusUnauthorized, "user update token permission deny", w)
		utils.Logger.Info("user update token permission deny")
		return 
	}
	username := req.Header["Username"]
	token := utils.MakeStringMd5(fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Int()))
	err := Dbcon.Model(&metaproxy.User{}).Where("user_name = ?", username).Update("user_token", token).Error
	if err != nil{
		SendReponseMsg(http.StatusInternalServerError, "user update token failed", w)
		utils.Logger.Error("user update token failed")
	}
	SendReponseMsg(http.StatusOK, token, w)
	utils.Logger.Info("user update token success")
}