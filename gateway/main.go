package main

import (
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"

	"github.com/lnsp/microlog/gateway/internal/email"
	"github.com/lnsp/microlog/gateway/internal/models"
	"github.com/lnsp/microlog/gateway/internal/router"
	"github.com/lnsp/microlog/gateway/pkg/utils"
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Hooks:     make(logrus.LevelHooks),
	Formatter: new(logrus.JSONFormatter),
	Level:     logrus.DebugLevel,
}

type specification struct {
	PublicAddr  string `default:"localhost:8080" desc:"Public address the server is reachable on"`
	Addr        string `default:":8080" desc:"Address the server is listening on"`
	Datasource  string `required:"true" desc:"Database file name"`
	Session     string `default:"secret" desc:"Shared session token secret"`
	Email       string `default:"secret" desc:"Shared email token secret"`
	SendgridKey string `envconfig:"SENDGRID_API_KEY" default:"" desc:"SendGrid API Key"`
	Minify      bool   `default:"false" desc:"Minify all responses"`
}

func main() {
	spec := &specification{}
	if err := envconfig.Process("micro", spec); err != nil {
		envconfig.Usage("micro", spec)
		return
	}
	dataSource, err := models.Open(spec.Datasource)
	if err != nil {
		log.WithFields(logrus.Fields{
			"datasource": spec.Datasource,
		}).Fatal("failed to open data source")
	}
	handler := router.New(router.Config{
		SessionSecret: []byte(spec.Session),
		EmailSecret:   []byte(spec.Email),
		EmailClient:   email.NewClient(dataSource, []byte(spec.Email), spec.SendgridKey),
		DataSource:    dataSource,
		PublicAddress: spec.PublicAddr,
		Minify:        true,
	})
	server := &http.Server{
		Handler:           logger(handler),
		Addr:              spec.Addr,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	log.WithFields(logrus.Fields{
		"addr": spec.Addr,
	}).Info("listening on address")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.WithError(err).Fatal("failed to listen")
	}
}

func logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		h.ServeHTTP(w, r)
		log.WithFields(logrus.Fields{
			"responseTime": time.Since(t).Seconds() * 1000.,
			"method":       r.Method,
			"path":         r.URL.Path,
			"remoteAddr":   utils.RemoteHost(r),
		}).Debug("handled request")
	})
}
