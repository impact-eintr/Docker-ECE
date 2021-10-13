package subsystems

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

type CpuSetSubSystem struct {
}

func (s *CpuSetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(cgroupPath, true); err == nil {
		if res.CpuSet != "" {
			if err := ioutil.WriteFile(
				path.Join(subsysCgroupPath, s.Name()+".cpus"),
				[]byte(res.CpuSet), 0644); err != nil {
				return fmt.Errorf("set cgroup cpu fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpuSetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpuSetSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(cgroupPath, false); err == nil {
		log.Println(cgroupPath, subsysCgroupPath, pid)
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

func (s *CpuSetSubSystem) Name() string {
	return "cpuset"
}
