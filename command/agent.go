package command

import (
	"github.com/alileza/potato/storage"
	"github.com/urfave/cli"
)

func NewAgentCommand() cli.Command {
	return cli.Command{
		Name:   "agent",
		Usage:  "Start potato agent",
		Action: startAgent,
	}
}

func startAgent(*cli.Context) error {
	storageServer := storage.NewStorage(&storage.Options{
		ListenAddress: "0.0.0.0:9005",
		StoragePath:   ".",
	})

	return storageServer.Serve()
}
