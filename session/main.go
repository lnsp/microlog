package main

import (
	"net"
	"time"

	health "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/kelseyhightower/envconfig"
	"github.com/lnsp/microlog/common/logger"
	"github.com/lnsp/microlog/session/api"
	"github.com/lnsp/microlog/session/internal/session"
	"google.golang.org/grpc"
)

var log = logger.New()

type specification struct {
	Secret        string `required:"true" desc:"Signing key for session tokens"`
	Redis         string `required:"true" desc:"Address for redis data store"`
	RedisPassword string `required:"true" desc:"Password for redis data store"`
	Addr          string `default:":8080" desc:"Address the service is listening on"`
	Expiration    string `default:"24h" desc:"Set expiration time"`
}

func main() {
	var spec specification
	if err := envconfig.Process("session", &spec); err != nil {
		envconfig.Usage("session", &spec)
		return
	}
	listener, err := net.Listen("tcp", spec.Addr)
	if err != nil {
		log.WithError(err).Fatal("could not setup networking")
	}
	expirationTime, err := time.ParseDuration(spec.Expiration)
	if err != nil {
		log.WithError(err).Fatal("bad expiration time format")
	}
	grpcServer := grpc.NewServer()
	sessionServer := session.NewServer(&session.Config{
		Secret:         []byte(spec.Secret),
		RedisAddr:      spec.Redis,
		RedisPassword:  spec.RedisPassword,
		ExpirationTime: expirationTime,
	})
	api.RegisterSessionServer(grpcServer, sessionServer)
	health.RegisterHealthServer(grpcServer, sessionServer.Health())
	if err := grpcServer.Serve(listener); err != nil {
		log.WithError(err).Fatal("failed to serve")
	}
}
