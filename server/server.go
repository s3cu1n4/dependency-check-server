package main

import (
	"dependency-check-server/common"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/s3cu1n4/logs/logs"
)

var (
	check         chan string
	dcpath        string
	notfirstCheck bool
)

func init() {
	if runtime.GOOS == "linux" {
		//docker image owasp/dependency-check path
		dcpath = "/usr/share/dependency-check/bin/dependency-check.sh"

	} else if runtime.GOOS == "darwin" {
		dcpath = "dependency-check"
	} else {
		logs.Fatal("the operating system not supported!")
		os.Exit(-1)
	}

}

type ReportJSON struct {
	ProjectInfo struct {
		ReportDate string `json:"reportDate"`
	}
	Dependencies []struct {
		FileName string `json:"fileName"`
		Md5      string `json:"md5"`
		Packages []struct {
			Id         string `json:"id"`
			Confidence string `json:"confidence"`
			Url        string `json:"url"`
		}
		VulnerabilityIds []struct {
			Id         string `json:"id"`
			Confidence string `json:"confidence"`
			Url        string `json:"url"`
		}
		Vulnerabilities []struct {
			Name     string `json:"name"`
			Severity string `json:"severity"`
			Cvssv3   struct {
				BaseScore    float64 `json:"baseScore"`
				AttackVector string  `json:"attackVector"`
			}
			Description string `json:"description"`
			References  []struct {
				Url string `json:"url"`
			}
		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	email := query.Get("email")
	project := query.Get("p")
	if project != "" {
		logs.Info(project)
		logs.Info(email)
		check <- project
		fmt.Fprintf(w, "start check")

	} else {
		fmt.Fprintf(w, "empty")
	}

}

func doCheck(project string) {
	prjRegexp := regexp.MustCompile(`(^.*?).jar`)
	ret := prjRegexp.FindStringSubmatch(project)
	// logs.Info(ret[1])
	if len(ret) == 2 {
		// logs.Info(PathExists(ret[1]))
		outpath := "./report/" + ret[1]
		projectpath := "/tmp/" + project

		var output []byte
		var cmd *exec.Cmd

		if !notfirstCheck {
			cmd = exec.Command(dcpath, "-o", outpath, "-f", "ALL", "-s", projectpath)
			logs.Info("服务启动后第一次检测，需要更新数据库，请耐心等待！")

		} else {
			cmd = exec.Command(dcpath, "-n", "-o", outpath, "-f", "ALL", "-s", projectpath)
		}

		logs.Info(cmd.Args)

		output, _ = cmd.Output()
		logs.Infof("check %s success", project)

		jsonFilePath := outpath + "/dependency-check-report.json"
		if common.Conf.DCServer.AliyunLog {
			decodeJson(jsonFilePath, project)
		}
		notfirstCheck = true
		logs.Info(string(output))
		return
	}
}

func decodeJson(jsonFile, project string) (reportjson ReportJSON) {
	f, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Println("read jsonfile err:", err)
	}

	er := json.Unmarshal(f, &reportjson)
	if er != nil {
		fmt.Println("jsondecode err", err)
	}
	generateLogs(reportjson, project)
	return

}

func PathExists(path string) (bool, error) {
	path = "./report/" + path
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
		return false, nil
	}
	return false, err
}

func generateLogs(reportjson ReportJSON, projectname string) []map[string]string {
	var retlist []map[string]string
	for _, dependencies := range reportjson.Dependencies {
		for k := range dependencies.Packages {
			for j := range dependencies.Vulnerabilities {
				alog := make(map[string]string)
				alog["projectName"] = projectname
				alog["depFileName"] = dependencies.FileName
				depname := strings.Split(dependencies.FileName, ":")
				if len(depname) > 1 {
					alog["depName"] = depname[1]
				}
				alog["reportDate"] = reportjson.ProjectInfo.ReportDate
				alog["depMd5"] = dependencies.Md5
				alog["depPackagesId"] = dependencies.Packages[k].Id
				alog["depPackagesConfidence"] = dependencies.Packages[k].Confidence
				alog["depPackagesUrl"] = dependencies.Packages[k].Url
				for l := range dependencies.VulnerabilityIds {
					alog["depVulIdsConfidence"] = dependencies.VulnerabilityIds[l].Confidence
					alog["depVulIdsId"] = dependencies.VulnerabilityIds[l].Id
					alog["depVulIdsUrl"] = dependencies.VulnerabilityIds[l].Url
				}
				alog["depVulName"] = dependencies.Vulnerabilities[j].Name

				alog["depVulSeverity"] = dependencies.Vulnerabilities[j].Severity
				alog["depVulCvssBaseScore"] = strconv.FormatFloat(dependencies.Vulnerabilities[j].Cvssv3.BaseScore, 'f', -1, 64)
				alog["depVulCvssAttackVector"] = dependencies.Vulnerabilities[j].Cvssv3.AttackVector
				alog["depVulDescription"] = dependencies.Vulnerabilities[j].Description
				retlist = append(retlist, alog)
			}
		}
	}
	common.Sendlog(retlist)
	return retlist
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		logs.Infof("FileName=[%s], FormName=[%s]\n", part.FileName(), part.FormName())
		dst, err := os.Create("/tmp/" + part.FileName())
		if err != nil {
			logs.Error("err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, part)
		check <- part.FileName()
	}
}

//在Linux环境下，定期更新漏洞库
func updateData() {
	var output []byte
	if runtime.GOOS == "linux" {
		for range time.Tick(2 * time.Hour) {
			time.Sleep(10 * time.Second)
			cmd := exec.Command(dcpath, "--updateonly")
			logs.Info("start update: ", cmd.Args)
			output, _ = cmd.Output()
			logs.Info(string(output))
		}
	}
}

func docheckchan() {
	logs.Info("start check chan")
	for project := range check {
		doCheck(project)

	}
}

func main() {
	check = make(chan string, 1000)
	var confpath string
	flag.StringVar(&confpath, "c", "conf/conf.yaml", "配置文件路径")
	flag.Parse()
	common.GetConfig(confpath, common.ServerConf)

	if !common.IsExist("./report") {
		common.CreateFile("./report")
	}
	logs.Info("start sucess")
	go updateData()
	go docheckchan()

	http.HandleFunc("/bbb", index)
	http.Handle("/report/", http.StripPrefix("/report/", http.FileServer(http.Dir("./report"))))
	http.HandleFunc("/uploadjar", uploadHandler)

	// 启动web服务，监听端口
	addr := fmt.Sprintf(":%s", common.Conf.DCServer.ListenPort)
	logs.Infof("Listen port is: %s", common.Conf.DCServer.ListenPort)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logs.Fatal("Listen port err: ", err)
	}

}
