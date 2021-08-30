package subsystems

type ResourceConfig struct {
	MemoryMax string
	CpuMax    string
	//CpuSet    string
}

type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	SubsystemsIns = []Subsystem{
		&MemorySubSystem{},
		&CpuSubSystem{},
		//&CpuSetSubSystem{},
	}
)
