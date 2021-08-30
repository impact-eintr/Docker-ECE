# 构建容器

## linux proc 文件系统
  linux下的`/proc`文件系统是由内核提供的，它其实不是一个真正的文件系统，只是包含了系统运行时的信息(比如系统内寻、mount设备信息、一些硬件配置等)，它只存在于内存中，而不占用外存空间。它以文件系统的方式，为访问内核数据的操作提供借口。实际上，很多系统工具都是简单地去读取这个文件系统的某个文件内容， 比如`lsmod` = `cat /procs/modules`
当遍历这个目录的时候，会发现很多数字，这些都是为每个进程创建的空间，数字就是它们的 PID 。

下面介绍几个比较重要的的部分

|PATH|作用|
|:--|:--:|
|/proc/N|PID为N的进程|
|/proc/N/cmdline|进程启动命令|
|/proc/N/cwd|链接带进程当前工作目录|
|/proc/N/environ|进程环境变量列表|
|/proc/N/exe|链接到进程的执行命令文件|
|/proc/N/fd|包含进程相关的所有文件描述符|
|/proc/N/maps|与进程相关的内存映射信息|
|/proc/N/mem|指代进程持有的内存，不可读|
|/proc/N/root|链接到进程的根目录|
|/proc/N/stat|进程的状态|
|/proc/N/statm|进程使用的内存状态|
|/proc/N/status|进程状态信息， 比stat/statm 更具可读性|
|/proc/self|链接到当前正在运行的进程|


``` go
const usage = `docker-ece is a simple container runtiome implementation.
The purpose of this project is to learn how docker works and how to
write a docker by ourselves. Enjoy it just for fun`

func main() {
	app := &cli.App{
		Name:  "docker-ece",
		Usage: usage,
	}

	app.Commands = []*cli.Command{
		&initCommand,
		&runCommand,
	}

	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

```

``` go
var runCommand = cli.Command{
	Name: "run",
	Usage: `Create  a container with namespace and cgroups limit
          mydocker run -ti [command ]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		&cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit",
		},
		&cli.StringFlag{
			Name:  "cpumax",
			Usage: "cpu limit",
		},
		//&cli.StringFlag{
		//	Name:  "cpuset",
		//	Usage: "cpuset limit",
		//},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return errors.New("Miss container command")
		}

		var cmdArr []string

		for _, arg := range ctx.Args().Slice() {
			cmdArr = append(cmdArr, arg)
		}

		tty := ctx.Bool("ti")
		resConf := &subsystems.ResourceConfig{
			MemoryMax: ctx.String("mem"),
			CpuMax:    ctx.String("cpumax"),
			//CpuSet:    ctx.String("cpuset"),
		}
		// Run 准备启动容器
		Run(tty, cmdArr, resConf)
		return nil
	},
}

var initCommand = cli.Command{
	Name: "init",
	Usage: `Init container process run user's process in container.
          Do not call it outside!`,
	Action: func(ctx *cli.Context) error {
		log.Infof("init comm on ")
		err := container.RunContainerInitProcess()
		return err
	},
}

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}

	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// use ece-cgroup as cgroup name
	cgroupManager := cgroups.NewCgroupManager("ece-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Apply(parent.Process.Pid)
	cgroupManager.Set(res)

	sendInitCommand(comArray, writePipe)
	parent.Wait()
	os.Exit(0)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

```


``` go
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe err %s", err)
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.ExtraFiles = []*os.File{readPipe}
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

```


``` go
func RunContainerInitProcess() error {
	comArray := readUserCommand()
	if comArray == nil || len(comArray) == 0 {
		return errors.New("Run container get user command error, cmdArray is nil")
	}

	defaultMountFlags := syscall.MS_NOEXEC | // 在本文件系统中不允许运行其他程序
		syscall.MS_NOSUID | // 在本系统中运行程序中，不允许set-user-ID set-grou-ID
		syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	path, err := exec.LookPath(comArray[0])
	if err != nil {
		log.Errorf("exec loop path error %v", err)
		return err
	}

	log.Infof("Find path %s", path)
	if err := syscall.Exec(path, comArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil

}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")

}

```


``` go
type CgroupManager struct {
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{Path: path}
}

// 将新的进程加入当前的cgroup中
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil

}

// 设置cgroup 资源限制
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Set(c.Path, res)
	}
	return nil

}

// 释放cgroup
func (c *CgroupManager) Destory() error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			log.Warn("remove cgroup fail %v", err)
		}
	}
	return nil

}

```


``` go
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

```


``` go
type CpuSubSystem struct {
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}

func (s *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CpuMax != "" {
			cpuMax, _ := strconv.Atoi(res.CpuMax)
			res.CpuMax = fmt.Sprintf("%d %d", cpuMax, 10000)
			if err := ioutil.WriteFile(
				path.Join(subsysCgroupPath, "cpu.max"),
				[]byte(res.CpuMax), 0644); err != nil {

				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
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

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

```


``` go
type MemorySubSystem struct {
}

func (s *MemorySubSystem) Name() string {
	return "memory"
}

func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.MemoryMax != "" {
			if err := ioutil.WriteFile(
				path.Join(subsysCgroupPath, "memory.max"),
				[]byte(res.MemoryMax), 0644); err != nil {

				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
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

func (s *MemorySubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

```

本章构造了一个简单的容器，具有基本的 amespace 隔离，并且确定了基本的开发架构，基本创建流程如图
![](/home/eintr/Projects/Docker-ECE/note/container/容器架构.png )

对于 Cgroups 通过这一章在容器上增加可配置的选项，可以实现对于容器可用资源的控制，最后，使用管道机制将用户输入的命令传递给容器初始化进程，实现了数据的传递
