package utils

import (
	"os"
	"fmt"
	"github.com/sirupsen/logrus"
)

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