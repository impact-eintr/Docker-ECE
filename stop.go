package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/impact-eintr/Docker-ECE/container"
	"github.com/sirupsen/logrus"
)

// 停止容器的Hook函数
func stopHook(containerName string) {
	// 获取容器信息
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("Get container %s info error %v", containerName, err)
		return
	}

	// 写入新的状态
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("Json marshal %s error %v", containerName, err)
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		logrus.Errorf("Write file %s error", configFilePath, err)
	}

	// 解除挂载
	if _, err := exec.Command("umount", containerInfo.RootUrl+"/merge").CombinedOutput(); err != nil {
		logrus.Errorf("Umount %s info error %v", containerInfo.RootUrl+"/merge", err)
		return
	}
}

func stopContainer(containerName string) {
	pid, err := GetContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("Get contaienr pid by name %s error %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logrus.Errorf("Conver pid from string to int error %v", err)
		return
	}
	// 杀进程
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		logrus.Errorf("Stop container %s error: %v", containerName, err)
	}
	stopHook(containerName)
}

func removeContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("Get container %s info error %v", containerName, err)
		return
	}
	if containerInfo.Status != container.STOP {
		logrus.Errorf("Couldn't remove running container")
		return
	}
	// 删除容器相关信息
	deleteContainerInfo(containerInfo.Id, containerName)
	// docker rm 的时候应该已经没有挂载
	container.DeleteWorkSpace(false, containerInfo.RootUrl, "")
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("Read file %s error %v", configFilePath, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		logrus.Errorf("GetContainerInfoByName unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}
