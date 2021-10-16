package main

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func commitContainer(containerName, imagesName string) {
	info, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get info of %s error %v", containerName, err)
	}
	mntURL := info.RootUrl + "/merge"
	imageTar := "/home/eintr/DockerImages/" + imagesName + ".tar"
	fmt.Printf("%s\n", imageTar)
	if _, err := exec.Command("tar", "-cvf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
