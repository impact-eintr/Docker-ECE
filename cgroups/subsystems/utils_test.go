package subsystems

import (
	"fmt"
	"testing"
)

func TestFindCgroupMountpoint(t *testing.T) {
	path, err := GetCgroupPath("memory", "memory.max", true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("cgroup path %v\n", path)
}
