package subsystems

type ResourceConfig struct {
	MemoryLimit string
	Cpu         string
	Cpuset      string
}

type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}
