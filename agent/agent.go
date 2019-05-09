package agent

import (
	"context"
	"net/http"
)

type Agent struct {
	options *Options

	httpServer *http.Server
}

type Options struct {
	ListenAddress string
}

func NewAgent(o *Options) *Agent {
	a := &Agent{
		httpServer: &http.Server{
			Addr: o.ListenAddress,
		},
	}

	return a
}

func (s *Agent) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Agent) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
