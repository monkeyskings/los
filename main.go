package main
import (
	"object-storage/controller"
	"object-storage/metaproxy"
	"object-storage/dataproxy"
	"os"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"object-storage/utils"
)

type Config struct {
	Global map[string]string
	Database map[string]string
	Storage map[string]string
}

var (
	configfile = "./config/object.conf"
)

func readyaml() (conf Config, err error) {
	yamlfile, err := ioutil.ReadFile(configfile)
	if err != nil {
		fmt.Println("open yaml file err : ", err)
		return conf, err
	}
	err = yaml.Unmarshal(yamlfile, &conf)
	if err != nil {
		fmt.Println("read yaml file err : ", err)
		return conf, err
	}
	return conf, nil
}

func main() {
	config, err := readyaml()
	if err != nil {
		os.Exit(-1)
	}
	logpath, ok := config.Global["logpath"]
	if ok == false{
		logpath = "./log/object/log"
	}
	err = utils.InitLog(logpath)
	if err != nil{
		os.Exit(-1)
	}
	utils.Logger.Info("start init env")
	if err = dataproxy.DataInit(config.Storage); err != nil{
		utils.Logger.Error("dataproxy init failed")
		os.Exit(-1)
	}
	utils.Logger.Info("dataproxy init success")
	if err = metaproxy.DbInitial(config.Database); err != nil{
		utils.Logger.Error("database init failed")
		os.Exit(-1)
	}
	utils.Logger.Info("metaproxy init success")

	db, err := metaproxy.DbOpen(config.Database)
	if err !=nil{
		utils.Logger.Error("database open failed")
		os.Exit(-1)
	}
	utils.Logger.Info("database open success")
	defer metaproxy.DbClose(db)
	
	err = controller.Start(db, config.Global)
	if err != nil{
		utils.Logger.Error("controller start failed")
		os.Exit(-1)
	}
	utils.Logger.Info("controller closed")
	//metaproxy.DbOpen()
}