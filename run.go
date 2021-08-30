package main

import (
	"errors"
	"os"
	"strings"

	"github.com/impact-eintr/Docker-ECE/cgroups"
	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
	"github.com/impact-eintr/Docker-ECE/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

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
