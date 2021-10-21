package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const usage = "Docker-ECE is a simple container runtime implementation"

func main() {

	app := cli.NewApp()
	app.Name = "Docker-ECE"
	app.Usage = usage

	log.Println(usage)

	app.Commands = []cli.Command{
		initCommand,
		reInitCommand,
		runCommand,
		commitCommand,
		listCommand,
		logCommand,
		execCommand,
		startCommand,
		stopCommand,
		removeCommand,
		networkCommand,
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
