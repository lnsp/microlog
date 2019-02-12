package common

import (
	"net"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

type logger struct {
	logrus.Logger
}

// WithRequest adds a remote host field to the log entry.
func (l *logger) WithRequest(req *http.Request) *logrus.Entry {
	return l.WithField("host", RemoteHost(req))
}

// RemoteHost returns the remote host IP.
func RemoteHost(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// Logger constructs a new logrus.Logger with additional helper methods.
func Logger() *logger {
	log := logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
	}
	if os.Getenv("DEBUG") != "" {
		log.Level = logrus.DebugLevel
		log.Formatter = new(logrus.TextFormatter)
	}
	return &logger{log}
}
