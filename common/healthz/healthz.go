// Package healthz provides a simple configurable healthz endpoint.
package healthz

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lnsp/microlog/common/logger"
)

var log = logger.New()

type Check func() error

type Endpoint struct {
	Checks []Check
}

func (e *Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Status     string   `json:"status"`
		StatusCode int      `json:"-"`
		Errors     []string `json:"errors"`
	}{
		"healthy", 200, make([]string, 0),
	}
	for _, check := range e.Checks {
		if err := check(); err != nil {
			resp.Status = "unhealthy"
			resp.StatusCode = http.StatusInternalServerError
			resp.Errors = append(resp.Errors, err.Error())
		}
	}
	data, _ := json.Marshal(resp)
	w.WriteHeader(resp.StatusCode)
	w.Write(data)
}

// New instantiates a new healthz endpoint.
func New(checks ...Check) http.Handler {
	return &Endpoint{checks}
}

// Run instantiates a new healthz endpoint in a new server and listens for connections.
func Run(addr string, checks ...Check) {
	endpoint := New(checks...)
	mux := http.NewServeMux()
	mux.Handle("/healthz", endpoint)
	server := http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.WithError(err).Fatal("health server failed to listen")
	}
}
