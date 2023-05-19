package util

import (
	"crypto/sha1"
	"fmt"
	"os"
)

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func FileSize(path string) uint64 {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return uint64(fileInfo.Size())
}

var sha1Size int64 = 1024

func QuickSHA1(path string) (string, error) {

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	finfo, _ := f.Stat()
	if finfo.Size() < sha1Size*2 {
		buf := make([]byte, sha1Size)
		f.Read(buf)
		s := sha1.New()
		s.Sum(buf)
		h := s.Sum(nil)
		return fmt.Sprintf("%x", h), nil
	}
	buf1 := make([]byte, sha1Size)
	buf2 := make([]byte, sha1Size)
	f.Read(buf1)
	f.Seek(-sha1Size, os.SEEK_END)
	f.Read(buf2)

	s := sha1.New()
	s.Write(buf1)
	s.Write(buf2)
	h := s.Sum(nil)
	return fmt.Sprintf("%x", h), nil
}
