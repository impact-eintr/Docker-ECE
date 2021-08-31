package common

import (
	"io/ioutil"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func GetAllFdsofProcess(path string) (fds []int) {
	result, err := ioutil.ReadDir(path)
	if err != nil {
		log.Error(err)
	}

	for _, fd := range result {
		f, err := strconv.ParseInt(fd.Name(), 10, 32)
		if err != nil {
			log.Errorf("Convert %s to int failed, Error Msg: %v", fd.Name(), err)
			return
		}
		fds = append(fds, int(f))
	}
	return
}
