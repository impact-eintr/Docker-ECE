package cgroups

import (
	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
	sub1 "github.com/impact-eintr/Docker-ECE/cgroups/subsystems/v1"
	sub2 "github.com/impact-eintr/Docker-ECE/cgroups/subsystems/v2"
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

func (c *CgroupManager) Apply2(pid int) error {
	sub2.SubsystemIns[0].Apply(c.Path, pid)
	return nil
}

func (c *CgroupManager) Set2(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range sub2.SubsystemIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

func (c *CgroupManager) Destroy2() error {
	if err := sub2.SubsystemIns[0].Remove(c.Path); err != nil {
		log.Warnf("remove cgroup fail %v", err)
	}
	return nil
}

func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range sub1.SubsystemIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range sub1.SubsystemIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range sub1.SubsystemIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			log.Warnf("remove cgroup fail %v", err)
		}
	}

	return nil
}
