package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/impact-eintr/Docker-ECE/cgroups"
	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
	"github.com/impact-eintr/Docker-ECE/container"
	log "github.com/sirupsen/logrus"
)

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/docker-ece/%s/"
	ConfigName          string = "config.json"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`         // 容器的init进程在宿主机上的 PID
	Id          string `json:"id"`          // 容器Id
	Name        string `json:"name"`        // 容器名
	Command     string `json:"command"`     // 容器内init运行命令
	CreatedTime string `json:"createdTime"` // 创建时间
	Status      string `json:"status"`      // 容器的状态
}

func Run(tty, version bool, comArray []string, res *subsystems.ResourceConfig,
	volume, containerName string) {

	containerInit, parent, writePipe := container.NewParentProcess(tty, volume)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("New parent process error")
	}

	// record container info
	containerName, err := recordContainerInfo(containerInit.Id,
		parent.Process.Pid, comArray, containerName)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}

	cgroupManager := cgroups.NewCgroupManager(containerInit.Id_base)
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

	if tty {
		parent.Wait()
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(containerInit.RootUrl, containerInit.MountUrl, volume)
	}
	//rootURL := "/var/lib/docker-ece"
	//mntURL := "/var/lib/docker-ece/merge"
	//container.DeleteWorkSpace(rootURL, mntURL, volume)
}

func recordContainerInfo(containerId string, containerPID int, commandArray []string,
	containerName string) (string, error) {

	//id := randStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	if containerName == "" {
		containerName = containerId
	}
	containerInfo := &container.ContainerInfo{
		Id:          containerId,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return "", err
	}

	return containerName, nil
}

func deleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}

//func randStringBytes(n int) string {
//	letterBytes := "1234567890"
//	rand.Seed(time.Now().UnixNano())
//	b := make([]byte, n)
//	for i := range b {
//		b[i] = letterBytes[rand.Intn(len(letterBytes))]
//	}
//	fmt.Println(string(b))
//	return string(b)
//}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
