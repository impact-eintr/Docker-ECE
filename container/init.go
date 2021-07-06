package container

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func RunContainerInitProcess() error {
	comArray := readUserCommand()
	if comArray == nil || len(comArray) == 0 {
		return errors.New("Run container get user command error, cmdArray is nil")
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	path, err := exec.LookPath(comArray[0])
	if err != nil {
		log.Errorf("exec loop path error %v", err)
		return err
	}

	log.Infof("Find path %s", path)
	if err := syscall.Exec("path", comArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil

}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")

}
