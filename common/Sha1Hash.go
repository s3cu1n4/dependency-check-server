package common

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/s3cu1n4/logs"
)

func computeSha1(r io.Reader) ([]byte, error) {
	h := sha1.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

type FileHash struct {
	Pre_hash string
	Hash     string
}

//ComputeFileSha1 计算一个文件的sha1
func ComputeFileSha1(filePtr *os.File, size int64) (hash FileHash, err error) {
	defer func() {
		if errr := recover(); errr != nil {
			logs.Error("ComputeAliFileSha1Error ", " error=", errr)
			// log.Println("ComputeAliFileSha1Error ", " error=", errr)
			err = errors.New("sha1 error")
		}
	}()
	hashbytes, err := computeSha1(filePtr)
	if err != nil {
		return hash, errors.New("sha1 error")
	}
	hash.Hash = fmt.Sprintf("%X", hashbytes)

	filePtr.Seek(0, 0)
	if size > 1024 {
		size = 1024
	}
	perbody := io.LimitReader(filePtr, size)
	prehashbytes, err := computeSha1(perbody)
	if err != nil {
		return hash, errors.New("sha1 error pre")
	}
	hash.Pre_hash = fmt.Sprintf("%X", prehashbytes)

	return hash, nil
}

func ComputeBuffSha1(filePtr *bytes.Reader, size int64) (hash FileHash, err error) {
	defer func() {
		if errr := recover(); errr != nil {
			// log.Println("ComputeAliBuffSha1Error ", " error=", errr)
			logs.Error("ComputeAliBuffSha1Error ", " error=", errr)

			err = errors.New("sha1 error")
		}
	}()

	//filePtr := bytes.NewReader(buff.Bytes())

	hashbytes, err := computeSha1(filePtr)
	if err != nil {
		return hash, errors.New("sha1 error")
	}
	hash.Hash = fmt.Sprintf("%X", hashbytes)

	filePtr.Seek(0, 0)
	if size > 1024 {
		size = 1024
	}
	perbody := io.LimitReader(filePtr, size)
	prehashbytes, err := computeSha1(perbody)
	if err != nil {
		return hash, errors.New("sha1 error pre")
	}
	hash.Pre_hash = fmt.Sprintf("%X", prehashbytes)

	return hash, nil
}
