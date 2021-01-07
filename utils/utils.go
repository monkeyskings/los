package utils

import (
	"os"
	"io"
	"fmt"
	"crypto/md5"
	"regexp"
	"net"
	"errors"
	"encoding/json"
	"io/ioutil"
	"github.com/sirupsen/logrus"
)

func MakeStringMd5(str string) string {
	m := md5.New()
	io.WriteString(m, str)
	md5str := fmt.Sprintf("%x", m.Sum(nil))
	return md5str
}

func CheckNameNormal(str string) bool {
    match, _ := regexp.MatchString("^[\\w\\d]+$", str) 
	return match
}

func CheckFileNameNormal(str string) bool {
	match, _ := regexp.MatchString("^[\\w\\d]+[.]*[\\w\\d]+$", str) 
	return match
}

func GetLocalIpaddr() (string , error){
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {

		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}

		}
	}
	return "", errors.New("can not find local ip address")
}

func ParseHttpBody(httpbody io.ReadCloser, v interface{}) error{
	body, err := ioutil.ReadAll(httpbody)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, v); err != nil {
		return err
	}
	return nil

}

var Logger = logrus.New()

func InitLog(filename string) error{
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0755)
	if err !=nil {
		fmt.Println("open logfile error : ", err)
		return err
	}
	Logger.SetOutput(logfile)
	return nil

}