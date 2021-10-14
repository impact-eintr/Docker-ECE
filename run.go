package main

import (
	"os"
	"strings"

	"github.com/impact-eintr/Docker-ECE/cgroups"
	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
	"github.com/impact-eintr/Docker-ECE/container"
	log "github.com/sirupsen/logrus"
)

func Run(tty, version bool, comArray []string, res *subsystems.ResourceConfig, volume string) {
	parent, writePipe := container.NewParentProcess(tty, volume)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("New parent process error")
	}

	cgroupManager := cgroups.NewCgroupManager("dockerece-cgroup")
	if version {
		defer cgroupManager.Destroy2()
		cgroupManager.Set2(res)
		cgroupManager.Apply2(parent.Process.Pid)
	} else {
		defer cgroupManager.Destroy()
		cgroupManager.Set(res)
		cgroupManager.Apply(parent.Process.Pid)
	}

	sendInitCommand(comArray, writePipe)
	parent.Wait()
	mntURL := "/home/eintr/Docker/merge/"
	rootURL := "/home/eintr/Docker/"
	container.DeleteWorkSpace(rootURL, mntURL, volume)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
