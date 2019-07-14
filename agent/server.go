package agent

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (a *Agent) startServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/fetch", a.fetchHandler())
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", a.Config.ServerPort),
		Handler: mux,
	}

	return srv.ListenAndServe()
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

		http.ServeFile(w, r, a.getFilePath(r.FormValue("event_id")))
	})
}
