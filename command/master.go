package command

import (
	"github.com/alileza/potato/storage"
	"github.com/urfave/cli"
)

func NewServerCommand() cli.Command {
	return cli.Command{
		Name:   "server",
		Usage:  "Start potato server",
		Action: startServer,
	}
}

func startServer(*cli.Context) error {
	storageServer := storage.NewStorage(&storage.Options{
		ListenAddress: "0.0.0.0:9005",
		StoragePath:   ".",
	})

	return storageServer.Serve()
}
