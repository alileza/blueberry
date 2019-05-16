package agent

import (
	"log"

	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:   "agent",
	Usage:  "Runs a Potato agent",
	Flags:  flags,
	Action: runCmd,
}

var flags = []cli.Flag{
	cli.BoolFlag{
		Name:  "client",
		Usage: "Enable client mode for the agent.",
	},
	cli.BoolFlag{
		Name:  "server",
		Usage: "Enable server mode for the agent.",
	},
	cli.StringSliceFlag{
		Name:  "join",
		Usage: "Address of an agent to join at start time. Can be specified multiple times.",
	},
}

func runCmd(ctx *cli.Context) error {
	if ctx.Bool("client") {
		log.Println("starting client")
	}
	if ctx.Bool("server") {
		log.Println("starting server")
	}

	agent := &Agent{}

	agent.options.Join = ctx.StringSlice("join")

	return agent.startAgent()
}
