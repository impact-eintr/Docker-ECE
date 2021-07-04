package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const usage = `docker-ece is a simple container runtiome implementation.
The purpose of this project is to learn how docker works and how to
write a docker by ourselves. Enjpy itm jut for fun`

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

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
