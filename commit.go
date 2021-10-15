package main

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func commitContainer(imagesName string) {
	mntURL := "/var/lib/docker-ece/merge"
	imageTar := "/home/eintr/DockerImages/" + imagesName + ".tar"
	fmt.Printf("%s\n", imageTar)
	if _, err := exec.Command("tar", "-cvf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
