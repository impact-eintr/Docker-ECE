package main

import (
	"fmt"

	"github.com/impact-eintr/Docker-ECE/cgroups/subsystems"
	"github.com/impact-eintr/Docker-ECE/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container",

	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		log.Info("command %s", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}

var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "cgroup2",
			Usage: "cgroup version 2",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpu",
			Usage: "cpu limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}

		tty := context.Bool("it")
		version := context.Bool("cgroup2")
		detach := context.Bool("d")
		if tty && detach {
			return fmt.Errorf("it and d paramter can not both provided")
		}
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			Cpu:         context.String("cpu"),
			Cpuset:      context.String("cpuset"),
		}
		volume := context.String("v")

		// TODO containerName network env port

		Run(tty, version, cmdArray, resConf, volume)
		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into images",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}
		imageName := context.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}
