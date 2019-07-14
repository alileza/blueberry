package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

func (a *Agent) Fetch(ctx context.Context, eventID string) error {
	agents, err := a.Consul.Agent().ServicesWithFilter("Service==" + a.Config.ServiceName)
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		for _, agent := range agents {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			req, err := http.NewRequest("GET", a.fetchURL(agent, eventID), nil)
			if err != nil {
				return err
			}

			if err := a.fetch(req.WithContext(ctx)); err != nil {
				a.Config.Logger.Printf("[%s] Failed to fetch from %s : %v\n", eventID, agent.Address, err)
				continue
			}

			a.Config.Logger.Printf("[%s] Fetched\n", eventID)
			return nil
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("[%s] Giving up to fetch from all agents\n", eventID)
}

func (a *Agent) fetchURL(svc *consulapi.AgentService, eventID string) string {
	u := &url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("%s:%d", svc.Address, svc.Port),
		Path:     "fetch",
		RawQuery: "event_id=" + eventID,
	}

	return u.String()
}

func (a *Agent) fetch(req *http.Request) error {
	resp, err := a.Config.ClientHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Service unavailable : %d", resp.StatusCode)
	}

	filepath := a.getFilePath(
		req.URL.Query().Get("event_id"),
	)

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
