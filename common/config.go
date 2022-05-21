package common

import (
	"os"
	"reflect"

	"github.com/s3cu1n4/logs"
	"github.com/spf13/viper"
)

var Conf = &Config{}

const (
	ClientConf = 1
	ServerConf = 2
)

type Config struct {
	DCServer DCServer `mapstructure:"DCServer"`
	Aliyun   Aliyun   `mapstructure:"aliyun"`
	Client   Client   `mapstructure:"client"`
}

type DCServer struct {
	ListenPort string `yaml:"listenPort"`
	AliyunLog  bool   `mapstructure:"aliyunlog"`
}

type Aliyun struct {
	Endpoint   string `mapstructure:"endpoint"`
	Ak         string `mapstructure:"ak"`
	Sk         string `mapstructure:"sk"`
	SlsProject string `mapstructure:"Slsproject"`
	Logstore   string `mapstructure:"Logstore"`
}

type Client struct {
	ServerAddr     string `mapstructure:"serveraddr"`
	MonitorJarPath string `mapstructure:"monitorjarpath"`
	JarTempPath    string `mapstructure:"jartemppath"`
}

func GetConfig(filename string, conftype int) {
	Conf = &Config{}
	viper.SetConfigType("yaml")
	viper.SetConfigFile(filename)
	err := viper.ReadInConfig()
	if err != nil {
		logs.Fatal(err)
		os.Exit(-1)
	}
	err = viper.Unmarshal(Conf)
	if err != nil {
		logs.Fatal(err.Error())
		os.Exit(-1)
	}
	if conftype == ServerConf {
		CheckServerConfig()
	} else if conftype == ClientConf {
		CheckClientConfig()

	}
}

func CheckServerConfig() {
	if Conf.DCServer.ListenPort == "" {
		logs.Fatalf("ListenPort is null")
		os.Exit(-1)

	}
	if Conf.DCServer.AliyunLog {
		t := reflect.TypeOf(Conf.Aliyun)
		v := reflect.ValueOf(Conf.Aliyun)
		for k := 0; k < t.NumField(); k++ {
			if v.Field(k).Interface() == "" {
				logs.Fatalf("阿里云配置项 %s 不能为空\n", t.Field(k).Tag.Get("mapstructure"))
				os.Exit(-1)

			}
		}
	}

}

func CheckClientConfig() {
	t := reflect.TypeOf(Conf.Client)
	v := reflect.ValueOf(Conf.Client)
	for k := 0; k < t.NumField(); k++ {
		if v.Field(k).Interface() == "" {
			logs.Fatalf("客户端配置 %s 不能为空\n", t.Field(k).Tag.Get("mapstructure"))
			os.Exit(-1)

		}
	}

}
