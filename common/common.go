package common

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/s3cu1n4/logs/logs"
	"pkg.re/essentialkaos/go-jar.v1"
)

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	return strings.Replace(dir, "\\", "/", -1)
}

func Md5sum(filepath string, check int) (md5str string, err error) {
	if check == 1 {
		f, err := os.Open(filepath)
		if err != nil {
			str1 := "Open err"
			return str1, err
		}
		defer f.Close()

		body, err := ioutil.ReadAll(f)
		if err != nil {
			str2 := "ioutil.ReadAll"
			return str2, err
		}
		md5str = fmt.Sprintf("%x", md5.Sum(body))
		runtime.GC()
		//return md5str, nil
	} else if check == 2 {
		data := []byte(filepath)
		has := md5.Sum(data)
		md5str = fmt.Sprintf("%x", has)
		//return md5str,nil
	}
	return md5str, nil
}

func Getfilepath(filepath string) (path string) {
	pathRegexp := regexp.MustCompile(`^.*\/(.*)`)
	params := pathRegexp.FindStringSubmatch(filepath)
	path = params[1]

	return
}

func CheckJar(fileName string) bool {
	logs.Infof("jar type check: %s", fileName)
	_, err := jar.ReadFile(fileName)
	return err == nil

}

func Getfileinfo(path string) (info os.FileInfo, err error) {
	info, err = os.Stat(path)
	if err != nil {
		logs.Error("get fileInfo failed:", err)
		return
	}
	return
}

func CopyFile(dstName, srcName string, n int64) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		logs.Error("copy file err:", err)
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logs.Error("copy file open err:", err)

		return
	}
	defer dst.Close()
	return io.CopyN(dst, src, n)
}

func CreateFile(filePath string) error {
	if !IsExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			logs.Error("create file err:", err)
			return err
		}
		return err

	}
	return nil
}

func IsExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		logs.Error("file not exist:", err)
		if os.IsExist(err) {
			return true
		} else {
			return false
		}

	}
	return true
}

func SendFile(server, filepath, filename string) error {
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	fileWriter, _ := bodyWriter.CreateFormFile("jarfiles", filename)

	file, _ := os.Open(filepath)
	defer file.Close()

	io.Copy(fileWriter, file)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(server, contentType, bodyBuffer)
	if err != nil {
		logs.Error("Send file error:", err)
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("read response err:", err)
		return err
	}
	// logs.Infof("Send to server: %s sucess,respinfo: %s ", server, string(sendresp))

	return err

}

func GzipDecode(in []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		var out []byte
		return out, err
	}
	defer reader.Close()

	return ioutil.ReadAll(reader)
}

func FormatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
