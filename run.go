package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/impact-eintr/Docker-ECE/cgroups"
	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
	"github.com/impact-eintr/Docker-ECE/container"
	log "github.com/sirupsen/logrus"
)

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("New parent process error")
	}

	cgroupManager := cgroups.NewCgroupManager("dockerece-cgroup")
	defer cgroupManager.Destroy()

	fmt.Println(res)
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(comArray, writePipe)
	parent.Wait()
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
