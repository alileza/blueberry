package agent

import (
	"testing"

	consulapi "github.com/hashicorp/consul/api"
)

func TestFetchURL(t *testing.T) {
	url := fetchURL(&consulapi.AgentService{
		Address: "localhost",
		Port:    6000,
	}, "event12332")

	if url != "http://localhost:6000/fetch??event_id=event12332" {
		t.Error("unexpected fetchURL")
	}
}
