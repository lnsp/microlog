package main

import (
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"

	"github.com/lnsp/microlog/internal/models"
	"github.com/lnsp/microlog/internal/router"
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Hooks:     make(logrus.LevelHooks),
	Formatter: new(logrus.TextFormatter),
	Level:     logrus.DebugLevel,
}

type specification struct {
	Addr       string `default:":8080" desc:"Address the server is listening on"`
	Datasource string `required:"true" desc:"Database file name"`
	Session    string `default:"secret" desc:"Shared session secret"`
}

func main() {
	spec := &specification{}
	if err := envconfig.Process("micro", spec); err != nil {
		envconfig.Usage("micro", spec)
		return
	}
	datasource, err := models.Open(spec.Datasource)
	if err != nil {
		log.Fatalln("Failed to open data source:", err)
	}
	handler := router.New(spec.Session, datasource)
	server := &http.Server{
		Handler:           logger(handler),
		Addr:              spec.Addr,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	log.Infoln("Listening on", spec.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln("Failed to listen:", err)
	}
}

func logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		h.ServeHTTP(w, r)
		log.Debugf("%.3fms %s %s", time.Since(t).Seconds()*1000., r.Method, r.URL.Path)
	})
}
