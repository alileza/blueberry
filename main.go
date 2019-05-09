package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/alileza/potato/command"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		command.NewAgentCommand(),
		command.NewServerCommand(),
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
