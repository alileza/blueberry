package agent

import (
	"log"

	"github.com/alileza/potato/storage"
	"github.com/urfave/cli"
)

type Command struct{}

func (a *Command) Usage() string {
	return "Runs a Potato agent"
}

func (a *Command) Flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "client",
			Usage: "Enable client mode for the agent.",
		},
		cli.BoolFlag{
			Name:  "server",
			Usage: "Enable server mode for the agent.",
		},
	}
}

func (a *Command) Run(ctx *cli.Context) error {
	if ctx.Bool("client") {
		log.Println("starting client")
	}
	if ctx.Bool("server") {
		log.Println("starting server")
	}
	storageServer := storage.NewStorage(&storage.Options{
		ListenAddress: "0.0.0.0:9005",
		StoragePath:   ".",
	})

	return storageServer.Serve()
}
