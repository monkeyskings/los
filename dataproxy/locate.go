package dataproxy

import (
	"mime/multipart"
	"io/ioutil"
	"net/http"
	"os"
	"io"
	"fmt"
	"strconv"
	"errors"
	"los/utils"
)
func LocateInit() error{
	storagepath, ok := StorageConf["path"]
	if ok == false{
		return errors.New("storageconf no path")
	}
	utils.Logger.Info("dataproxy locate start init")
	info, err := os.Stat(storagepath)
	if err != nil{
		utils.Logger.Info("dataproxy locate init to mkdir : ", storagepath)
		return os.Mkdir(storagepath, os.ModePerm)
	}
	if info.IsDir() == false{
		return errors.New("storage path is file")
	}
	return nil
}

func LocateCreate(filename string, file multipart.File)error {
	storagepath, ok := StorageConf["path"]
	if ok == false{
		return errors.New("storageconf no path")
	}
	utils.Logger.Info("dataproxy locate ready to create local file : ", filename)
	data, err:=ioutil.ReadAll(file)
	if err!=nil{
		return err
	}
	err = ioutil.WriteFile(storagepath + filename, data, 0666)
	utils.Logger.Info("dataproxy locate finish write local file : ", filename)
	return err
}

func LocateRead(filename string, writer http.ResponseWriter) error{
	storagepath, ok := StorageConf["path"]
	if ok == false{
		return errors.New("storageconf no path")
	}
	filename = fmt.Sprintf("%s%s", storagepath, filename)
	utils.Logger.Info("dataproxy locate ready to read local file : ", filename)
	file, err := os.Open(filename)
	if err != nil{
		utils.Logger.Error("dataproxy locate file open error : ", filename)
		return err
	}
	defer file.Close()

	fileheader := make([]byte, 512)
	file.Read(fileheader)

	filestat, err := file.Stat()
	if err !=nil {
		utils.Logger.Error("dataproxy locate file stat error : ", filename)
		return err
	}
	
	writer.Header().Set("Content-Disposition", "attachment; filename=" + filename)
	writer.Header().Set("Content-Type", http.DetectContentType(fileheader))
	writer.Header().Set("Content-Length", strconv.FormatInt(filestat.Size(), 10))
	utils.Logger.Info("dataproxy locate finish to write header : ", filename)

	_, err = file.Seek(0, 0)

	if err != nil {
		utils.Logger.Error("dataproxy locate file seek error : ", filename)
		return err
	}

	_, err = io.Copy(writer, file)
	utils.Logger.Info("dataproxy locate finish to download : ", filename)
	return err
}