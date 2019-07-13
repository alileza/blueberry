package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/mholt/archiver"
	"github.com/oklog/run"
)

func main() {
	consulConf := api.DefaultConfig()

	consulConf.Address = "consul:8500"

	consulClient, err := consulapi.NewClient(consulConf)
	if err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "server":
		if err := server(consulClient); err != nil {
			panic(err)
		}
	case "distribute":
		if err := distribute(consulClient, os.Args[2:]...); err != nil {
			panic(err)
		}
	}
}

func distribute(consulClient *consulapi.Client, filenames ...string) error {
	eventID, writeMeta, err := consulClient.Event().Fire(&consulapi.UserEvent{
		Name: "potato",
	}, &consulapi.WriteOptions{})
	if err != nil {
		return err
	}

	err = archiver.Archive(filenames, "storage/"+eventID+".zip")
	if err != nil {
		return err
	}

	fmt.Println(eventID)
	fmt.Printf("%+v", writeMeta)
	return nil
}

func server(consulClient *consulapi.Client) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	err = consulClient.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		ID:      hostname,
		Name:    "potato",
		Address: hostname,
		Port:    6000,
	})
	if err != nil {
		return err
	}

	var (
		g       run.Group
		errChan = make(chan error)
	)

	g.Add(func() error {
		return watchEvents(consulClient)
	}, func(err error) {
		errChan <- err
	})

	g.Add(func() error {
		srv := &http.Server{
			Addr:    "0.0.0.0:6000",
			Handler: handler(),
		}
		return srv.ListenAndServe()
	}, func(err error) {
		errChan <- err
	})
	if err := g.Run(); err != nil {
		return err
	}
	return <-errChan
}

func handler() http.Handler {
	var inFlight int64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&inFlight, 1)
		defer atomic.AddInt64(&inFlight, -1)

		if inFlight >= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		time.Sleep(time.Second)
		http.ServeFile(w, r, "storage/"+r.URL.Path)
	})
}

func watchEvents(consulClient *consulapi.Client) error {
	receivedEvents := make(map[string][]byte)

	for range time.Tick(time.Second) {
		events, meta, err := consulClient.Event().List("potato", &consulapi.QueryOptions{})
		if err != nil {
			return err
		}

		for _, event := range events {
			if _, ok := receivedEvents[event.ID]; ok {
				continue
			}

			receivedEvents[event.ID] = event.Payload

			fmt.Printf("[%s] Downloading... %+v\n", event.ID, event)
			potatoes, err := getPotatoes(consulClient)
			if err != nil {
				return err
			}
		download:
			var downloaded bool
			for _, potato := range potatoes {
				url := potato.Address + "/" + event.ID + ".zip"
				if err := downloadFile(url, "storage/"+event.ID+".zip"); err != nil {
					fmt.Printf("[%s] Failed to download from %s : %v\n", event.ID, url, err)
					continue
				}
				downloaded = true
				fmt.Printf("[%s] Downloaded\n", event.ID)
			}
			if !downloaded {
				goto download
			}
		}
		fmt.Printf("request_time=%s event_list=%d received_event=%+v\n", meta.RequestTime, len(events), len(receivedEvents))
	}
	return nil
}

type Potato struct {
	Address string
}

func getPotatoes(consulClient *consulapi.Client) ([]Potato, error) {
	potatoServices, err := consulClient.Agent().ServicesWithFilter("Service==potato")
	if err != nil {
		return nil, err
	}
	var potatoes []Potato
	for _, p := range potatoServices {
		potatoes = append(potatoes, Potato{
			Address: fmt.Sprintf("http://%s:%d", p.Address, p.Port),
		})
	}
	return potatoes, nil
}

func downloadFile(url string, target string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Service unavailable")
	}

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
