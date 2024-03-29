package main

import (
	"net/http"
	"time"

	"github.com/lnsp/microlog/common/logger"
	"github.com/lnsp/microlog/gateway/internal/session"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/lnsp/microlog/gateway/internal/email"
	"github.com/lnsp/microlog/gateway/internal/models"
	"github.com/lnsp/microlog/gateway/internal/router"
)

var log = logger.New()

type specification struct {
	PublicAddr     string `default:"localhost:8080" desc:"Public address the server is reachable on"`
	Addr           string `default:":8080" desc:"Address the server is listening on"`
	Datasource     string `required:"true" desc:"Database file name"`
	Minify         bool   `default:"false" desc:"Minify all responses"`
	EmailService   string `default:"mail:8080" desc:"Email service host"`
	SessionService string `default:"session:8080" desc:"Session service host"`
	CsrfAuthKey    string `default:"csrf-auth-key" desc:"CSRF validation key"`
	CsrfSecure     bool   `default:"true" desc:"CSRF HTTPS only"`
}

func main() {
	spec := &specification{}
	if err := envconfig.Process("micro", spec); err != nil {
		envconfig.Usage("micro", spec)
		return
	}
	dataSource, err := models.Open(spec.Datasource)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"datasource": spec.Datasource,
		}).Fatal("failed to open data source")
	}
	handler := router.New(router.Config{
		EmailClient:   email.NewClient(dataSource, spec.EmailService),
		SessionClient: session.NewClient(dataSource, spec.SessionService),
		DataSource:    dataSource,
		PublicAddress: spec.PublicAddr,
		Minify:        spec.Minify,
		CsrfAuthKey:   []byte(spec.CsrfAuthKey),
		CsrfSecure:    spec.CsrfSecure,
	})
	server := &http.Server{
		Handler:           log.Middleware(handler),
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
