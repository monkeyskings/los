package dataproxy
import (
	"mime/multipart"
	"net/http"
	"errors"
	"los/utils"
)

var StorageConf map[string]string
func DataInit(dataconf map[string]string) error{
	StorageConf = dataconf
	storagetype, ok := dataconf["mode"]
	if ok == false{
		return errors.New("storageconf no mode")
	}
	utils.Logger.Info("dataproxy start init")
	if storagetype == "locate"{
		return LocateInit()
	}
	return nil
}

func DataCreate(filename string, file multipart.File) error{
	storagetype, ok := StorageConf["mode"]
	if ok == false{
		return errors.New("storageconf no mode")
	}
	if storagetype == "locate"{
		return LocateCreate(filename, file)
	}
	return nil
}

func DataRead(filename string, writer http.ResponseWriter) error{
	storagetype, ok := StorageConf["mode"]
	if ok == false{
		return errors.New("storageconf no mode")
	}
	if storagetype == "locate"{
		return LocateRead(filename, writer)
	}
	return nil
}