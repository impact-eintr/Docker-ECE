# Docker-ECE

《自己动手写docker》的学习笔记

## 基础知识
### Linux Namespace
linux namespace是kernel的一个功能，他可以隔离一系列的系统资源(PID UserID Network)
### Linux Cgroups
Linux Cgroups提供了对一组进程及将来子进程的资源限制控制和统计的能力，这些资源包括CPU、内存、存储、网络等。通过Cgroups,可以方便地限制某个进程的资源占用，并且可以实时地监控进程的监控与统计信息

``` go
package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	cmd := exec.Command("bash")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			//syscall.CLONE_NEWUSER,
			syscall.CLONE_NEWNET,
	}

	//cmd.SysProcAttr.Credential = &syscall.Credential{
	//	Uid: uint32(1),
	//	Gid: uint32(1),
	//}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
	os.Exit(-1)
}

```

#### Cgroup中的3个组件
- **cgroup** 一个cgroup中包含一组进程，并且可以在这个cgroup上增加Linux subsystem的各种参数配置，将一组进程和一组subsystem的系统参数关联起来
- **subsystem** 
  - blkio 块设备
  - cpu cpu调度策略
  - cpuaccta cgroup 中进程的cpu占用
  - cpuset 多核机器上设置cgroup中进程可以使用的cpu与内存
  - devices 控制cgroup中进程对设备的访问
  - freezera suspend resume
  - memory 控制内存占用
  - net_cls 将cgroup中进程产生的网络包进行分类，以便tc分流
  - net_prio 设置cgroup中进程产生的网络包的优先级
  - ns 使cgroup中的进程在新的nameSpace中fork新进程时创建出一个新的cgroup 这个cgroup包含新的nameSpace中的进程
  
- **hierarchy** 将一组cgroup串成一个树状的结构，一个这样的树便是一个hierarchy

#### 关系
- 系统在创建了新的 hierarchy 之后,系统中所有的进程都会加入这个 hierarchy 的cgroup根节点,这个 cgroup 根节点是 hierarchy 默认创建的。
- 一个 subsystem 只能附加到 一 个 hierarchy 上面。
- 一个 hierarchy 可以附加多个 subsystem 。
- 一个进程可以作为多个 cgroup 的成员,但是这些 cgroup 必须在不同的 hierarchy 中 。
- 一个进程fork出子进程时,子进程是和父进程在同一个 cgroup 中的,也可以根据需要将其移动到其他 cgroup 中 。


