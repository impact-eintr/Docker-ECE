package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var Root string

func GetCgroupPath(cgroupPath string, autoCreate bool) (string, error) {

	var cgroupRoot string
	if cgroupPath != "" {
		Root = FindCgroupMountpoint()
		cgroupRoot = FindAbsoluteCgroupMountpoint()

	} else {
		cgroupRoot = Root
		fmt.Println(cgroupRoot)
	}

	_, err := os.Stat(path.Join(cgroupRoot, cgroupPath))
	if err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), os.ModePerm); err != nil {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil

	} else {
		return "", fmt.Errorf("cgroup path error %v", err)

	}
}
func FindAbsoluteCgroupMountpoint() string {
	return "/sys/fs/cgroup"
}

func FindCgroupMountpoint() string {
	f, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return ""
	}
	defer f.Close()

	rawPath, err := ioutil.ReadAll(f)
	if err != nil {
		return ""
	}
	arr := strings.Split(string(rawPath), ":")
	path := arr[len(arr)-1]

	return "/sys/fs/cgroup" + string(path[:len(path)-1])

}
