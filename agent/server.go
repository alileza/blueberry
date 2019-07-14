package agent

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (a *Agent) startServer(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", a.listHandler())
	mux.HandleFunc("/fetch", a.fetchHandler())
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	srv := &http.Server{
		Addr:     fmt.Sprintf("%s:%d", a.Config.ServerAddress, a.Config.ServerPort),
		Handler:  mux,
		ErrorLog: a.Config.Logger,
	}

	errChan := make(chan error)
	go func() {
		errChan <- srv.ListenAndServe()
	}()

	select {
	case e := <-errChan:
		return e
	case <-ctx.Done():
		return srv.Shutdown(ctx)
	}
}

func (a *Agent) fetchHandler() http.HandlerFunc {
	var inFlight int64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&inFlight, 1)
		defer atomic.AddInt64(&inFlight, -1)

		if inFlight > a.Config.ServerMaxConn {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		time.Sleep(time.Second)

		filepath := a.getFilePath(r.FormValue("event_id"))

		_, err := os.Stat(filepath)
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.ServeFile(w, r, filepath)
	})
}

func (a *Agent) listHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files, err := ioutil.ReadDir(a.Config.ServerStoragePath)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
		for _, file := range files {
			fmt.Fprintf(w, "%s\t%s\n", file.ModTime().Format(time.RFC3339), file.Name())
		}
	})
}
