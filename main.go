package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const cgroupMemoryHierarchyMount = "/sys/fs/cgroup"

func main() {

	if os.Args[0] == "/proc/self/exe" {
		// 容器进程
		fmt.Printf("fork again! current pid %d\n", syscall.Getpid())
		cmd := exec.Command("sh", "-c", `/bin/stress --vm-bytes 1024m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
	}

	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalln(err)
		os.Exit(1)
	} else {
		// 得到fork出来进程映射在外部命名空间的pid
		fmt.Printf("forked [%v]\n", cmd.Process.Pid)
		//在系统默认创建挂载了 memory subsystem 的 Hierarchy 上创建 cgroup
		os.Mkdir(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit"), 0755)

		//将容器进程加入到这个 cgroup 中
		ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "cgroup.procs"),
			[]byte(strconv.Itoa(cmd.Process.Pid)), 0644)

		//限制 cgroup 进程使用
		ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "cpu.max"),
			[]byte("50000 100000"), 0644)

		ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "memory.max"),
			[]byte("200m"), 0644)

		cmd.Process.Wait()

	}

}
