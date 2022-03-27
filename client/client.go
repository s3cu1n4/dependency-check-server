package main

import (
	"dependency-check-server/common"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/s3cu1n4/logs/logs"
	"gopkg.in/fsnotify.v1"
)

var (
	jarPATH string
	prjLog  sync.Map
	jarTemp sync.Map
)

type Watch struct {
	watch *fsnotify.Watcher
}

//监控目录
func (w *Watch) watchDir(dir string) {
	logs.Info("Start jar monitor from base dir:", dir)
	go func() {
		for {
			select {
			case ev := <-w.watch.Events:
				{
					if ev.Op&fsnotify.Create == fsnotify.Create {
						fi, err := os.Stat(ev.Name)
						if err == nil && fi.IsDir() {
							w.watch.Add(ev.Name)
						}
					}

					if ev.Op&fsnotify.Write == fsnotify.Write {

						if checkType(ev.Name) {
							go SendJar2Server(ev.Name)
						}
					}
				}
			case err := <-w.watch.Errors:
				{
					logs.Error("watch error : ", err)
					continue
				}

			}
		}
	}()
	//通过Walk来遍历目录下的所有子目录
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		//这里判断是否为目录，只需监控目录即可
		//目录下的文件也在监控范围内，不需要我们一个一个加
		if info.IsDir() {
			path, err := filepath.Abs(path)
			if err != nil {
				logs.Error(err)
				return err
			}
			err = w.watch.Add(path)
			if err != nil {
				logs.Error(err)
				return err
			}
		}
		return nil
	})
}

func SendJar2Server(path string) {
	jarname := common.Getfilepath(path)

	now := time.Now().UnixNano()
	if _, ok := jarTemp.LoadOrStore(jarname, now); ok {
		jarTemp.Store(jarname, now)
	}

	time.Sleep(1 * time.Second)
	if val, ok := jarTemp.Load(jarname); ok {
		if val == now {
			//  1 秒内未更新过
			if common.CheckJar(path) {
				md5, err := common.Md5sum(path, 1)
				if err != nil {
					logs.Error("get file hash err:", err)
					return
				}
				if _, ok := prjLog.Load(jarname); ok {
					// hash 重复，不需要重复检测
					logs.Info("jar包hash值重复:", jarname)
					return
					// jarTemp.Store(jarname, now)
				} else {
					// prjLog[jarname] = md5
					//未检测过的hash值，需要重新检测
					dstPath := jarPATH + jarname
					info, err := common.Getfileinfo(path)
					if err != nil {
						logs.Error("get file info err", err)
						return
					}
					if info.Size() > 1024*500 {
						n, err := common.CopyFile(dstPath, path, info.Size())
						if err != nil {
							logs.Error("copy file err", err)
							return
						}
						err = common.SendFile(common.Conf.Client.ServerAddr, dstPath, jarname)
						if err != nil {
							return
						}
						logs.Infof("Filename: %s FileSize: %s send sucess", jarname, common.FormatFileSize(n))
					}
					prjLog.Store(jarname, md5)
				}

			}

		} else {
			return
		}
	}

}

func checkType(path string) bool {
	matched, _ := regexp.MatchString(`^.*\.jar$`, path)
	if matched {
		return true
	} else {
		return false
	}
}

func checkServerStatus() bool {

	url := strings.Replace(common.Conf.Client.ServerAddr, "uploadjar", "bbb", 1)
	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		logs.Fatal(err.Error())
		return false
	}
	logs.Infof("连接服务端成功，%s %s  ", common.Conf.Client.ServerAddr, resp.Status)
	return true

}

func main() {
	var confpath string
	flag.StringVar(&confpath, "c", "conf.yaml", "配置文件路径")
	flag.Parse()

	common.GetConfig(confpath, common.ClientConf)
	if !common.IsExist(common.Conf.Client.JarTempPath) {
		common.CreateFile(common.Conf.Client.JarTempPath)
	}
	if !common.IsExist(common.Conf.Client.MonitorJarPath) {
		logs.Error("Monitor path is not found:", common.Conf.Client.MonitorJarPath)
		os.Exit(-1)
	}
	jarPATH = common.Conf.Client.JarTempPath
	if !checkServerStatus() {
		logs.Errorf("连接服务端 %s 失败，请检查与服务端的网络是否能正常访问。", common.Conf.Client.ServerAddr)
		os.Exit(-1)

	}

	watch, _ := fsnotify.NewWatcher()
	w := Watch{
		watch: watch,
	}

	w.watchDir(common.Conf.Client.MonitorJarPath)
	logs.Infof("dependency check server: %s Monitor path is: %s ", common.Conf.Client.ServerAddr, common.Conf.Client.MonitorJarPath)
	logs.Info("启动jar包监控服务完成")
	select {}

}