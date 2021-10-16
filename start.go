package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/impact-eintr/Docker-ECE/container"
	"github.com/sirupsen/logrus"
)

// start 指令
func startContainer(containerName string) {
	//init, err := GetContainerInitByName(containerName)
	//if err != nil {
	//	logrus.Errorf("Get contaienr pid by name %s error %v", containerName, err)
	//	return
	//}
	// TODO  重新挂载 mnt
	// container.CreateMountPoint(init.RootUrl)

	// 把新的容器信息写回配置文件
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("Get container %s info error %v", containerName, err)
		return
	}

	containerInfo.Status = container.RUNNING
	containerInfo.Pid = " " // TODO 获取新的PID

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
	fmt.Println(containerName)
}

func GetContainerInitByName(containerName string) (*container.ContainerInit, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &container.ContainerInit{}, nil
}
