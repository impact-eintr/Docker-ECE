package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/impact-eintr/Docker-ECE/cgroups"
	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
	"github.com/impact-eintr/Docker-ECE/container"
	"github.com/impact-eintr/Docker-ECE/network"
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
	ImageUrl    string `json:"imageUrl"`    // 容器挂载镜像 这个其实应该可以省略
	RootUrl     string `json:"rootUrl"`     // 容器挂载目录集的根目录
}

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig,
	volume, imageName, containerName string, envSlice []string,
	nw string, portmapping []string) {

	// containerInit 包含容器初始化时需要记录的一些信息
	containerInit, parent, writePipe := container.NewParentProcess(tty, imageName, volume, envSlice)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("New parent process error: %v", err)
	}

	// record container info
	containerName, err := recordContainerInfo(containerInit, parent.Process.Pid,
		containerName, comArray)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}

	// 开启cgroup
	// 检查版本
	_, err = exec.Command("grep", "cgroup2", "/proc/filesystems").CombinedOutput()
	if err != nil {
		// cgroup 1
		cgroupManager := cgroups.NewCgroupManager("")
		if tty {
			defer cgroupManager.Destroy()
		}
		cgroupManager.Set(res)
		cgroupManager.Apply(parent.Process.Pid)

	} else {
		// cgroup 2
		cgroupManager := cgroups.NewCgroupManager(containerInit.Id_base)
		if tty {
			defer cgroupManager.Destroy2()
		}
		cgroupManager.Set2(res)
		cgroupManager.Apply2(parent.Process.Pid)
	}

	if nw != "" {
		// config container network
		network.Init()
		containerInfo := &container.ContainerInfo{
			Id:          containerInit.Id,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portmapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			log.Errorf("Error Connetc Newwork %v", err)
			return
		}
	}

	sendInitCommand(comArray, writePipe)

	if tty {
		parent.Wait()
		// 停止容器
		stopHook(containerName)
	}
}

func recordContainerInfo(containerInit *container.ContainerInit, containerPID int,
	containerName string, commandArray []string) (string, error) {

	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, " ")
	fmt.Println(command)
	var flag bool
	if containerName == "" {
		containerName = containerInit.Id
		flag = true
	}
	containerInfo := &container.ContainerInfo{
		Id:          containerInit.Id,
		Pid:         strconv.Itoa(containerPID),
		Name:        containerName,
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		ImageUrl:    containerInit.ImageUrl,
		RootUrl:     containerInit.RootUrl,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerInit.Id)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return "", err
	}

	if !flag {
		linkUrl := fmt.Sprintf(container.DefaultInfoLocation[:len(container.DefaultInfoLocation)-1],
			containerName)
		if err := os.Symlink(dirUrl, linkUrl); err != nil {
			log.Errorf("Link error %s error %v", dirUrl, err)
			return "", err
		}
	}

	fileName := dirUrl + container.ConfigName
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

func deleteContainerInfo(containerId, containerName string) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Errorf("Remove dir %s error %v", dirUrl, err)
	}
	linkUrl := fmt.Sprintf(container.DefaultInfoLocation[:len(container.DefaultInfoLocation)-1],
		containerName)
	if err := os.RemoveAll(linkUrl); err != nil {
		log.Errorf("Remove dir %s error %v", dirUrl, err)
	}
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
