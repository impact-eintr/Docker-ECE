package v1

import (
	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
)

var (
	SubsystemIns = []subsystems.Subsystem{
		&CpusetSubSystem{},
		&CpuSubSystem{},
		&MemorySubSystem{},
	}
)
