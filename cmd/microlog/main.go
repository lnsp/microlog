package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/logging"
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
	PublicAddr string `default:"localhost:8080" desc:"Public address the server is reachable on"`
	Addr       string `default:":8080" desc:"Address the server is listening on"`
	Datasource string `required:"true" desc:"Database file name"`
	Session    string `default:"secret" desc:"Shared session token secret"`
	Email      string `default:"secret" desc:"Shared email token secret"`
	Minify     bool   `default:"false" desc:"Minify all responses"`
	ProjectID  string `default:"microlog-213919" desc:"Google Cloud Project ID"`
}

func main() {
	spec := &specification{}
	if err := envconfig.Process("micro", spec); err != nil {
		envconfig.Usage("micro", spec)
		return
	}
	dataSource, err := models.Open(spec.Datasource)
	if err != nil {
		log.Fatalln("Failed to open data source:", err)
	}
	handler := router.New(router.Config{
		SessionSecret: []byte(spec.Session),
		EmailSecret:   []byte(spec.Email),
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
	log.Infoln("Listening on", spec.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln("Failed to listen:", err)
	}
}

func logger(gclouID string, h http.Handler) http.Handler {
	client, err := logging.NewClient(context.Background(), gcloudID)
	if err != nil {
		log.Fatalln("Failed to setup logging:", err)
	}
	logger := client.Logger("microlog").StandardLogger(logging.Debug)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		h.ServeHTTP(w, r)
		logger.Printf("%.3fms %s %s", time.Since(t).Seconds()*1000., r.Method, r.URL.Path)
	})
}
