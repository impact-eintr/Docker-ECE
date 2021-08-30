package container

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func RunContainerInitProcess() error {
	comArray := readUserCommand()
	fmt.Println(comArray)
	if comArray == nil || len(comArray) == 0 {
		return errors.New("Run container get user command error, cmdArray is nil")
	}

	setUpMount()

	path, err := exec.LookPath(comArray[0])
	if err != nil {
		log.Errorf("exec loop path error %v", err)
		return err
	}

	log.Infof("Find path %s", path)
	if err := syscall.Exec(path, comArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil

}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")

}

// 挂载点
func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return
	}
	log.Infof("Current location is %s", pwd)

	if err := pivotRoot(pwd); err != nil {
		log.Errorf("Mount rootfs error %v", err)
		return

	}

	defaultMountFlags := syscall.MS_NOEXEC | // 在本文件系统中不允许运行其他程序
		syscall.MS_NOSUID | // 在本系统中运行程序中，不允许set-user-ID set-grou-ID
		syscall.MS_NODEV // mount的系统的默认参数
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")

}

func pivotRoot(root string) error {
	// 为了使当前root的老root和新root不在同一个文件系统下，我们把root重新mount了一遍，
	// bind mount 是把相同的内容换了一个挂载点的挂载方法
	//if err := syscall.Mount(root, root, "bind", syscall.MS_PRIVATE, ""); err != nil {
	//	return fmt.Errorf("Mount rootfs to itself error: %v", err)
	//}

	exec.Command("mount", "--make-rprivate", "/").CombinedOutput()

	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if _, err := os.Stat(pivotDir); err != nil {
		if err := os.Mkdir(pivotDir, 0777); err != nil {
			return err
		}
	}

	// pivot_root到新的rootfs 现在老的 old_root挂载在rootfs/.pivot_root
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		fmt.Println(root, pivotDir)
		return fmt.Errorf("pivot_root %v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// 删除临时文件夹
	return os.Remove(pivotDir)
}
