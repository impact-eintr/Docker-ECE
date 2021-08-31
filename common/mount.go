package common

import (
	"log"
	"strings"
	"syscall"
)

var (
	Mount  = syscall.Mount
	Umount = syscall.Unmount
)

// 获取所有的挂载点
func FindMountPoint() ([]string, error) {
	v := Must2(Exec("mount"))
	switch tp := v.(type) {
	case string:
		log.Println(tp)
		return parseMountInfo(tp), nil
	default:
		return nil, LogAndErrorf("Unexpected type: %T", tp)
	}
}

func parseMountInfo(info string) (result []string) {
	arrays := strings.Split(info, "\n")
	root := DockerRoot

	for _, val := range arrays {
		point := strings.Split(val, " ")[2]
		if idx := strings.Index(point, root); idx != -1 {
			result = append(result, point)
		}
	}
	return result

}
