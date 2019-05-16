package agent

import (
	"log"

	"github.com/alileza/potato/storage"
	"github.com/hashicorp/serf/serf"
	"github.com/oklog/run"
)

type Agent struct {
	options       Options
	serfClient    *serf.Serf
	eventHandlers map[string]func(serf.Event)
}

type Options struct {
	Join []string
}

func (a *Agent) registerHandler(eventName string, eh func(serf.Event)) {
	if a.eventHandlers == nil {
		a.eventHandlers = make(map[string]func(serf.Event))
	}

	a.eventHandlers[eventName] = eh
}

func (a *Agent) startAgent() error {
	var g run.Group
	g.Add(func() error {
		conf := serf.DefaultConfig()

		eventCh := make(chan serf.Event, 64)
		conf.EventCh = eventCh

		s, err := serf.Create(conf)
		if err != nil {
			return err
		}

		a.serfClient = s

		if len(a.options.Join) > 0 {
			a.serfClient.Join(a.options.Join, false)
		}

		go func() {
			for {
				ev := <-eventCh

				if ev.EventType() != serf.EventUser {
					continue
				}
				eh, ok := a.eventHandlers[ev.String()]
				if !ok {
					log.Println("[WARN] Unhandled event type")
				}
				eh(ev)
			}
		}()

		<-a.serfClient.ShutdownCh()
		return nil
	}, func(err error) {
		log.Println(err)
		a.serfClient.Shutdown()
	})

	g.Add(func() error {
		storageServer := storage.NewStorage(&storage.Options{
			ListenAddress: "0.0.0.0:9006",
			StoragePath:   ".",
		})

		return storageServer.Serve()
	}, func(err error) {
		log.Println(err)
	})

	return g.Run()
}

// serfEventHandler is used to handle events from the serf cluster
func (s *Agent) serfEventHandler() {

}
