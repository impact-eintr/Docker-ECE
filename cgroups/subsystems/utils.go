package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountpoint()

	_, err := os.Stat(path.Join(cgroupRoot, cgroupPath))
	if err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err == nil {
			} else {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil

	} else {
		return "", fmt.Errorf("cgroup path error %v", err)

	}
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
