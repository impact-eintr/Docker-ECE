package subsystems

import (
	"os"
	"testing"
	"time"
)

func TestCpuCgroup(t *testing.T) {
	memSubSys := CpuSubSystem{}
	resConfig := ResourceConfig{
		CpuMax: "50000 100000",
	}
	testCgroup := "testcpulimit"

	if err := memSubSys.Set(testCgroup, &resConfig); err != nil {
		t.Fatalf("cgroup fail %v", err)
	}

	if err := memSubSys.Apply(testCgroup, os.Getpid()); err != nil {
		t.Fatalf("cgroup Apply %v", err)
	}

	go func() {
		i := 0
		for {
			i++
		}
	}()

	time.Sleep(5 * time.Second)

	//将进程移回到根Cgroup节点
	if err := memSubSys.Apply("", os.Getpid()); err != nil {
		t.Fatalf("cgroup Apply %v", err)
	}

	go func() {
		i := 0
		for {
			i++
		}
	}()

	//if err := memSubSys.Remove(testCgroup); err != nil {
	//	t.Fatalf("cgroup remove %v", err)
	//}
	for {
		select {}
	}
}
