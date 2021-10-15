package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Error("New pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}

	//mntURL := "/home/eintr/Docker/merge"
	//rootURL := "/home/eintr/Docker"
	imageURL := "/home/eintr/DockerImages"
	rootURL := "/var/lib/docker-ece"
	mntURL := "/var/lib/docker-ece/merge"
	NewWorkSpace(imageURL, rootURL, mntURL, volume)
	cmd.Dir = mntURL
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

func NewWorkSpace(imageURL, rootURL, mntURL, volume string) {
	if err := os.Mkdir(rootURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", rootURL, err)
	}
	CreateLowerLayer(imageURL, rootURL)
	CreateUpperLayer(rootURL)
	CreateWorkDir(rootURL)

	CreateMountPoint(rootURL, mntURL) // 创建merge层
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(rootURL, mntURL, volumeURLs)
			log.Infof("%q", volumeURLs)
		} else {
			log.Infof("Vplume parameter input is not correct.")
		}
	}
}

func volumeUrlExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

func MountVolume(rootURL string, mntURL string, volumeURLs []string) {
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		log.Infof("Mkdir parent dir %s error. %v", parentUrl, err)
	}

	containerUrl := volumeURLs[1]
	containerVolumeURL := mntURL + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Infof("Mkdir container dir %s error. %v", containerVolumeURL, err)
	}

	cmd := exec.Command("mount", "--bind", parentUrl, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func CreateLowerLayer(imageURL, rootURL string) {
	busyboxURL := rootURL + "/busybox"
	busyboxTarURL := imageURL + "/busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	fmt.Println(busyboxTarURL, busyboxURL)
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL,
			"-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("unTar dir %s error. %v", busyboxTarURL, err)
		}
	}
}

func CreateUpperLayer(rootURL string) {
	upperURL := rootURL + "/upper"
	if err := os.Mkdir(upperURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", upperURL, err)
	}
}

func CreateWorkDir(rootURL string) {
	workURL := rootURL + "/work"
	if err := os.Mkdir(workURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", workURL, err)
	}
}

func CreateMountPoint(rootURL string, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error.%v", mntURL, err)
	}
	dirs := "lowerdir=" + rootURL + "/busybox" +
		",upperdir=" + rootURL + "/upper" +
		",workdir=" + rootURL + "/work"

	fmt.Println(dirs)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(rootURL, mntURL, volumeURLs)
		} else {
			DeleteMountPoint(rootURL, mntURL)
		}
	} else {
		DeleteMountPoint(rootURL, mntURL)
	}
	DeleteWriteLayer(rootURL)
}

func DeleteWriteLayer(rootURL string) {
	workURL := rootURL + "/work"
	if err := os.RemoveAll(workURL); err != nil {
		log.Errorf("Remove dir %s error %v", workURL, err)
	}
	upperURL := rootURL + "/upper"
	if err := os.RemoveAll(upperURL); err != nil {
		log.Errorf("Remove dir %s error %v", upperURL, err)
	}
}

func DeleteMountPointWithVolume(rootURL string, mntURL string, volumeURLs []string) {
	containerUrl := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}

	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}

	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
