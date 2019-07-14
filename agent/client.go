package agent

import (
	"errors"
	"io"
	"net/http"
	"os"
)

func (a *Agent) Fetch(eventID string) error {
	var downloaded bool
	potatoes, err := a.Consul.Agent().ServicesWithFilter("Service==" + a.Config.ServiceName)
	if err != nil {
		return err
	}

download:
	for _, potato := range potatoes {
		url := potato.Address + "/fetch?event_id=" + eventID
		if err := downloadFile(url, a.getFilePath(eventID)); err != nil {
			a.Config.Logger.Printf("[%s] Failed to download from %s : %v\n", eventID, url, err)
			continue
		}
		downloaded = true
		a.Config.Logger.Printf("[%s] Downloaded\n", eventID)
	}
	if !downloaded {
		goto download
	}
	return nil
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
