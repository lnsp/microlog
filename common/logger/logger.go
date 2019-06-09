package logger

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	logrus.Logger
}

// WithRequest adds a remote host field to the log entry.
func (l *Logger) WithRequest(req *http.Request) *logrus.Entry {
	return l.WithField("host", RemoteHost(req))
}

// Middleware constructs a new http handler that logs incoming http requests.
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		next.ServeHTTP(w, r)
		l.WithRequest(r).WithFields(logrus.Fields{
			"responseTime": time.Since(t).Seconds() * 1000.,
			"method":       r.Method,
			"path":         r.URL.Path,
		}).Debug("handled request")
	})
}

// RemoteHost returns the remote host IP.
func RemoteHost(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// New constructs a new logrus.Logger with additional helper methods.
func New() *Logger {
	log := logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
	}
	if os.Getenv("DEBUG") != "" {
		log.Level = logrus.DebugLevel
		log.Formatter = new(logrus.TextFormatter)
	}
	return &Logger{log}
}
