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

	"github.com/s3cu1n4/logs"
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

// 检测jar包是否写完，
func SendJar2Server(path string) {

	jarname := common.Getfilepath(path)

	now := time.Now().UnixNano()

	if _, ok := jarTemp.LoadOrStore(jarname, now); ok {
		jarTemp.Store(jarname, now)
		return
	} else {
		//第一次收到文件写入信号
		logs.Infof("start check jar: %s", jarname)
	}

	ticker := time.NewTicker(time.Second * 1)
	count := 0
	var filePtr *os.File

	for range ticker.C {
		//写入超时判断，超过60秒为超时，超时后该jar包不检测
		if count > 60 {
			return
		}
		count++
		if !common.CheckFileIsOpen(path) {
			logs.Infof("%s write sucess, start check", jarname)
			jarTemp.Delete(jarname)
			if common.CheckJar(path) {
				filePtr, err := os.OpenFile(path, os.O_RDONLY, 0666)
				if err != nil {
					logs.Error("open file err:", err)
					return
				}

				info, err := common.Getfileinfo(path)
				if err != nil {
					return
				}

				// jar包小于500KB，不检测
				if info.Size() < 1024*500 {
					logs.Infof("jar包:%s , %s 小于500KB,不检测", info.Name(), common.FormatFileSize(info.Size()))
					return
				}

				hash, err := common.ComputeFileSha1(filePtr, info.Size())
				filePtr.Close()
				if err != nil {
					logs.Error("Compute FileSha1 err:", err)
					return
				}

				// hash 重复，不需要重复检测
				if val, ok := prjLog.Load(jarname); ok {
					if val.(string) == hash.Hash {
						logs.Infof("jar: %s hash: %s 值重复:", jarname, hash.Hash)
						return
					}
				}
				// prjLog[jarname] = md5
				// 未检测过的hash值，需要重新检测
				dstPath := jarPATH + jarname
				if err != nil {
					logs.Error("get file info err", err)
					return
				}

				n, err := common.CopyFile(dstPath, path, info.Size())
				if err != nil {
					return
				}

				err = common.SendFile(common.Conf.Client.ServerAddr, dstPath, jarname)
				if err != nil {
					return
				}

				logs.Infof("Filename: %s FileSize: %s file hash: %s send sucess", jarname, common.FormatFileSize(n), hash.Hash)
				prjLog.Store(jarname, hash.Hash)
				return

			} else {
				logs.Infof(" %s is not jar", jarname)
				return
			}

		}
	}
	defer filePtr.Close()

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
