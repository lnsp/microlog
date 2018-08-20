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
	server := &http.Server{
		Handler:           router.New(spec.Session, datasource),
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
