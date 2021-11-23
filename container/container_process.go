package container

import (
	"encoding/base32"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/docker-ece/%s/"
	ConfigName          string = "config.json"
	ContainerLogFile    string = "container.log"
)

type ContainerInfo struct {
	Pid         string   `json:"pid"`         // 容器的init进程在宿主机上的 PID
	Id          string   `json:"id"`          // 容器Id
	Name        string   `json:"name"`        // 容器名
	Command     string   `json:"command"`     // 容器内init运行命令
	CreatedTime string   `json:"createdTime"` // 创建时间
	Status      string   `json:"status"`      // 容器的状态
	ImageUrl    string   `json:"imageUrl"`    // 容器挂载镜像 这个其实应该可以省略
	RootUrl     string   `json:"rootUrl"`     // 容器挂载目录集的根目录
	PortMapping []string `json:"portmapping"` // 端口映射
}

type ContainerInit struct {
	Id       string
	Id_base  string
	ImageUrl string
	RootUrl  string
}

func NewId() string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Encode(b []byte) string {
	return base32.StdEncoding.EncodeToString(b)
}

func NewParentProcess(tty bool, imageName, volume string, envSlice []string) (*ContainerInit, *exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Error("New pipe error %v", err)
		return nil, nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	// 从镜像构造容器
	id := NewId()
	id_base := Encode([]byte(id))
	rootURL := "/var/lib/docker-ece/" + id_base
	mntURL := "/var/lib/docker-ece/" + id_base + "/merge"
	var imageURL string
	if imageName != "" {
		imageURL = "Images/" + imageName + ".tar"
	} else {
		imageURL = "Images/busybox.tar"
	}

	cmd.ExtraFiles = []*os.File{readPipe}
	NewWorkSpace(imageURL, rootURL, mntURL, volume)
	cmd.Env = append(os.Environ(), envSlice...)
	cmd.Dir = mntURL

	// 构造日志输出
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// 生成容器对应日志目录
		dirURL := fmt.Sprintf(DefaultInfoLocation, id)
		if err := os.MkdirAll(dirURL, 0622); err != nil && !os.IsExist(err) {
			log.Errorf("NewParentProcess mkdir %serror %v", dirURL, err)
			return nil, nil, nil
		}
		stdLogFilePath := dirURL + ContainerLogFile
		stdLogFile, err := os.OpenFile(stdLogFilePath,
			os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)
		if err != nil {
			log.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
			return nil, nil, nil
		}
		cmd.Stdout = stdLogFile
		cmd.Stderr = stdLogFile
	}
	return &ContainerInit{id, id_base, imageURL, rootURL}, cmd, writePipe
}

// TODO 暂时不支持volume
func ReNewParentProcess(info *ContainerInfo, volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Error("New pipe error %v", err)
		return nil, nil
	}

	// 重新启动进程
	cmd := exec.Command("/proc/self/exe", "reinit")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	// 重新挂载容器读写层
	mntURL := info.RootUrl + "/merge"
	NewWorkSpace(info.ImageUrl, info.RootUrl, mntURL, volume)

	cmd.Dir = mntURL
	cmd.ExtraFiles = []*os.File{readPipe}

	// 生成容器对应日志目录
	dirURL := fmt.Sprintf(DefaultInfoLocation, info.Id)
	if err := os.MkdirAll(dirURL, 0622); err != nil && !os.IsExist(err) {
		log.Errorf("NewParentProcess mkdir %serror %v", dirURL, err)
		return nil, nil
	}
	stdLogFilePath := dirURL + ContainerLogFile
	stdLogFile, err := os.OpenFile(stdLogFilePath,
		os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)
	if err != nil {
		log.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
		return nil, nil
	}
	//cmd.Stdin = os.Stdin
	cmd.Stdout = stdLogFile
	cmd.Stderr = stdLogFile

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

	CreateRootDir(rootURL)
	CreateLowerLayer(imageURL, rootURL)
	CreateUpperLayer(rootURL)
	CreateWorkDir(rootURL)

	CreateMountPoint(imageURL, rootURL) // 创建merge层
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
	if err := os.MkdirAll(parentUrl, 0777); err != nil {
		log.Infof("Mkdir parent dir %s error. %v", parentUrl, err)
	}

	containerUrl := volumeURLs[1]
	containerVolumeURL := mntURL + containerUrl
	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
		log.Infof("Mkdir container dir %s error. %v", containerVolumeURL, err)
	}

	cmd := exec.Command("mount", "--bind", parentUrl, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func CreateRootDir(rootURL string) {
	exist, err := PathExists(rootURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", rootURL, err)
	}
	if !exist {
		if err := os.MkdirAll(rootURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", rootURL, err)
		}
	}
}

func CreateLowerLayer(imageURL, rootURL string) {
	if imageURL == "" {
		return
	}
	var busyboxURL, busyboxTarURL string
	busyboxURL = rootURL + "/lower"
	busyboxTarURL = imageURL

	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if !exist {
		if err := os.MkdirAll(busyboxURL, 0777); err != nil {
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
	exist, err := PathExists(upperURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", upperURL, err)
	}
	if !exist {
		if err := os.MkdirAll(upperURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", upperURL, err)
		}
	}
}

func CreateWorkDir(rootURL string) {
	workURL := rootURL + "/work"
	exist, err := PathExists(workURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", workURL, err)
	}
	if !exist {
		if err := os.MkdirAll(workURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", workURL, err)
		}
	}
}

func CreateMountPoint(imageURL string, rootURL string) {
	mntURL := rootURL + "/merge"
	exist, err := PathExists(mntURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", mntURL, err)
	}
	if !exist {
		if err := os.MkdirAll(mntURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", mntURL, err)
		}
	}
	var dirs string
	if imageURL != "" {
		dirs = "lowerdir=" + rootURL + "/lower" +
			",upperdir=" + rootURL + "/upper" +
			",workdir=" + rootURL + "/work"
	} else {
		fakeRoot := "/home/eintr/Downloads/root"
		dirs = "lowerdir=" + fakeRoot +
			",upperdir=" + rootURL + "/upper" +
			",workdir=" + rootURL + "/work"
	}

	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func DeleteWorkSpace(umount bool, rootURL, volume string) {
	if umount {
		if volume != "" {
			volumeURLs := volumeUrlExtract(volume)
			length := len(volumeURLs)
			if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
				DeleteMountPointWithVolume(rootURL, volumeURLs)
			} else {
				DeleteMountPoint(rootURL)
			}
		} else {
			DeleteMountPoint(rootURL)
		}
	}
	DeleteWriteLayer(rootURL)
}

func DeleteWriteLayer(rootURL string) {
	if err := os.RemoveAll(rootURL); err != nil {
		log.Errorf("Remove dir %s error %v", rootURL, err)
	}
}

func DeleteMountPointWithVolume(rootURL string, volumeURLs []string) {
	mntURL := rootURL + "/merge"
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

	// 这里不用删除挂载目录 删除读写层时会一并删除
	//if err := os.RemoveAll(mntURL); err != nil {
	//	log.Errorf("Remove dir %s error %v", mntURL, err)
	//}
}

func DeleteMountPoint(rootURL string) {
	mntURL := rootURL + "/merge"
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
	// 这里不用删除挂载目录 删除读写层是会一并删除
	//if err := os.RemoveAll(mntURL); err != nil {
	//	log.Errorf("Remove dir %s error %v", mntURL, err)
	//}
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
