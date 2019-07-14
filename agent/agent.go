package agent

import (
	"log"
	"os"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/mholt/archiver"
)

type Config struct {
	ID              string
	Hostname        string
	ServiceName     string
	CompressionType string

	ConsulConfig *consulapi.Config

	Logger *log.Logger

	ServerPort        int
	ServerMaxConn     int64
	ServerStoragePath string

	EventName          string
	EventWatchInterval time.Duration
}

func DefaultConfig() *Config {
	hostname, _ := os.Hostname()
	return &Config{
		ID:              hostname,
		Hostname:        hostname,
		ServiceName:     "potato",
		CompressionType: "zip",

		ConsulConfig: consulapi.DefaultConfig(),
		Logger:       log.New(os.Stdout, "potato: ", 0),

		ServerPort:        6000,
		ServerMaxConn:     2,
		ServerStoragePath: "storage",

		EventName:          "potato",
		EventWatchInterval: time.Second,
	}
}

type Agent struct {
	Config *Config
	Consul *consulapi.Client
}

func NewAgent(conf *Config) (*Agent, error) {
	consulClient, err := consulapi.NewClient(conf.ConsulConfig)
	if err != nil {
		return nil, err
	}

	return &Agent{
		Config: conf,
		Consul: consulClient,
	}, nil
}

func (a *Agent) Run() error {
	err := a.Consul.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		ID:      a.Config.ID,
		Name:    a.Config.ServiceName,
		Address: a.Config.ID,
		Port:    a.Config.ServerPort,
	})
	if err != nil {
		return err
	}

	go a.watch()

	return a.startServer()
}

func (a *Agent) watch() error {
	receivedEvents := make(map[string][]byte)
	eventChan := make(chan *consulapi.UserEvent)

	ticker := time.NewTicker(a.Config.EventWatchInterval)
	defer ticker.Stop()

	for range ticker.C {
		events, _, err := a.Consul.Event().List(a.Config.EventName, &consulapi.QueryOptions{})
		if err != nil {
			return err
		}

		for _, event := range events {
			if _, ok := receivedEvents[event.ID]; ok {
				continue
			}

			receivedEvents[event.ID] = event.Payload

			eventChan <- event
		}
	}

	return nil
}

func (a *Agent) Distribute(filenames ...string) error {
	eventID, _, err := a.Consul.Event().Fire(&consulapi.UserEvent{
		Name: a.Config.EventName,
	}, &consulapi.WriteOptions{})
	if err != nil {
		return err
	}

	err = archiver.Archive(filenames, a.getFilePath(eventID))
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) getFilePath(eventID string) string {
	return a.Config.ServerStoragePath + "/" + eventID + "." + a.Config.CompressionType
}
