package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/alileza/potato/command/agent"
)

func main() {
	app := cli.NewApp()

	commands := map[string]Command{
		"agent": &agent.Command{},
	}

	for name, cmd := range commands {
		app.Commands = append(app.Commands, cli.Command{
			Name:   name,
			Usage:  cmd.Usage(),
			Action: cmd.Run,
			Flags:  cmd.Flags(),
		})
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

type Command interface {
	Usage() string
	Run(*cli.Context) error
	Flags() []cli.Flag
}
