package common

import (
	"io"
	"os"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetRandomNumber() string {
	s, err := uuid.NewRandom()
	if err != nil {
		log.Warnf("UUID error.%v", err)
	}
	return s.String()
}

func NameExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsExist(err) {
		return true
	}
	return false
}

func ReadFromFd(fd uintptr) (string, error) {
	f := os.NewFile(fd, "cmdInit")
	b, err := io.ReadAll(f)
	if err != nil {
		return "", LogAndErrorsf("Failed to read from fd %d, error: %v", fd, err)
	}
	return string(b), nil
}

const (
	quiet     = "-q"
	directory = "-p"
	verbose   = "-v"
)

func DownLoadFromUrl(url string, savedDir string) {
	if _, err := Exec("wget", verbose, directory, savedDir, url); err != nil {
		log.Error(err)
	}
}
