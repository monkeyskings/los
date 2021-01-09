package utils

import (
	"io"
	"fmt"
	"crypto/md5"
	"regexp"
	"net"
	"errors"
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

