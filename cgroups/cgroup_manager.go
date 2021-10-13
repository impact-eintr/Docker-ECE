package cgroups

import (
	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
	log "github.com/sirupsen/logrus"
)

type CgroupManager struct {
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (c *CgroupManager) Apply(pid int) error {
	//for _, subSysIns := range subsystems.SubsystemIns {
	//	subSysIns.Apply(c.Path, pid)
	//}
	subsystems.SubsystemIns[0].Apply(c.Path, pid)
	return nil
}

func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

func (c *CgroupManager) Destroy() error {
	//for _, subSysIns := range subsystems.SubsystemIns {
	//	if err := subSysIns.Remove(c.Path); err != nil {
	//		log.Warnf("remove cgroup fail %v", err)
	//	}
	//}

	if err := subsystems.SubsystemIns[0].Remove(c.Path); err != nil {
		log.Warnf("remove cgroup fail %v", err)
	}
	return nil
}
