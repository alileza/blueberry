package main

import (
	"os"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/alileza/potato/agent"
)

func main() {
	consulConf := consulapi.DefaultConfig()
	consulConf.Address = "consul:8500"

	conf := agent.DefaultConfig()
	conf.ConsulConfig = consulConf

	a, err := agent.NewAgent(conf)
	if err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "agent":
		if err := a.Run(); err != nil {
			panic(err)
		}
	case "distribute":
		if err := a.Distribute(os.Args[2:]...); err != nil {
			panic(err)
		}
	}
}
