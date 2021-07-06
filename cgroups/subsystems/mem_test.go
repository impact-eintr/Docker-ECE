package subsystems

import (
	"os"
	"path"
	"testing"
)

func TestMemoryCgroup(t *testing.T) {
	memSubSys := MemorySubSystem{}
	resConfig := ResourceConfig{
		MemoryMax: "1000m",
	}
	testCgroup := "testmemorylimit"

	stat, err := os.Stat(path.Join("/sys/fs/cgroup", testCgroup))
	t.Logf("cgroup stats: %+v, %v", stat, err)

	if err := memSubSys.Set(testCgroup, &resConfig); err != nil {
		t.Fatalf("cgroup fail %v", err)
	}
	//if err := memSubSys.Apply(testCgroup, os.Getpid()); err != nil {
	//	t.Fatalf("cgroup Apply %v", err)
	//}
	////将进程移回到根Cgroup节点
	//if err := memSubSys.Apply("", os.Getpid()); err != nil {
	//	t.Fatalf("cgroup Apply %v", err)
	//}

	//if err := memSubSys.Remove(testCgroup); err != nil {
	//	t.Fatalf("cgroup remove %v", err)
	//}
}
