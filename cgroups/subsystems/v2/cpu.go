package v2

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
)

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Set(cgroupPath string, res *subsystems.ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(cgroupPath, true); err == nil {
		if res.Cpu != "" {
			per, _ := strconv.Atoi(res.Cpu)
			if per < 0 || per > 100 {
				return fmt.Errorf("set cgroup cpu fail: resource")
			}
			res.Cpu = fmt.Sprintf("%d %d", per*1000, 100000)
			if err := ioutil.WriteFile(
				path.Join(subsysCgroupPath, s.Name()+".max"),
				[]byte(res.Cpu), 0644); err != nil {
				return fmt.Errorf("set cgroup cpu fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(cgroupPath, false); err == nil {
		log.Println(cgroupPath, subsysCgroupPath, pid)
		// 注意WriteFile 是截断写
		if err := ioutil.WriteFile(
			path.Join(subsysCgroupPath, "cgroup.procs"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {

			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}
