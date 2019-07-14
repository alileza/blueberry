package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/mholt/archiver"
	"github.com/oklog/run"
)

type Config struct {
	ID              string
	Hostname        string
	ServiceName     string
	CompressionType string

	ConsulConfig *consulapi.Config

	Logger *log.Logger

	ClientHTTP *http.Client

	ServerAddress     string
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
		Logger:       log.New(os.Stdout, "["+time.Now().Format(time.RFC3339)+"] ", 0),

		ClientHTTP: &http.Client{},

		ServerAddress:     "0.0.0.0",
		ServerPort:        6000,
		ServerMaxConn:     2,
		ServerStoragePath: "/storage",

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

	defer a.Consul.Agent().ServiceDeregister(a.Config.ID)

	var g run.Group
	{
		// Termination handler.
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		g.Add(
			func() error {
				select {
				case <-term:
					a.Config.Logger.Print("Received SIGTERM, exiting gracefully...")
				}
				return nil
			},
			func(err error) {
			},
		)
	}
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			err := a.watch(ctx)
			a.Config.Logger.Println("watch: Stopped")
			return err
		}, func(err error) {
			a.Config.Logger.Println("watch: Stopping")
			cancel()
		})
	}
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			err := a.startServer(ctx)
			a.Config.Logger.Printf("server: Stopped")
			return err
		}, func(err error) {
			a.Config.Logger.Printf("server: Stopping")
			cancel()
		})
	}
	fmt.Printf("Potato agent configuration:\n\n\tNode Name: %s\n\tStorage Path: %s\n\tListen Address: 0.0.0.0:%d\n\n", a.Config.ID, a.Config.ServerStoragePath, a.Config.ServerPort)
	return g.Run()
}

func (a *Agent) watch(ctx context.Context) error {
	receivedEvents := make(map[string][]byte)

	ticker := time.NewTicker(a.Config.EventWatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			events, _, err := a.Consul.Event().List(a.Config.EventName, &consulapi.QueryOptions{})
			if err != nil {
				return err
			}
			a.Config.Logger.Printf("event: %d", len(events))

			for _, event := range events {
				if _, ok := receivedEvents[event.ID]; ok {
					continue
				}
				a.Config.Logger.Printf("event: Incoming %s", event.ID)
				receivedEvents[event.ID] = event.Payload

				if err := a.Fetch(ctx, event.ID); err != nil {
					a.Config.Logger.Printf("event: Failed to fetch %s", event.ID)
				} else {
					a.Config.Logger.Printf("event: Fetched %s", event.ID)
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
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
