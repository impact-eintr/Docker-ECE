package v1

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
)

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Set(cgroupPath string, res *subsystems.ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.Cpu != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpu.shares"), []byte(res.Cpu), 0644); err != nil {
				return fmt.Errorf("set cgroup cpu share fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
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
