package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/lnsp/microlog/session/api"
	"github.com/lnsp/microlog/session/pkg/session"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

type specification struct {
	Secret        string `required:"true" desc:"Signing key for session tokens"`
	Redis         string `required:"true" desc:"Address for redis data store"`
	RedisPassword string `required:"true" desc:"Password for redis data store"`
	Addr          string `default:":8080" desc:"Address the service is listening on"`
}

func main() {
	var spec specification
	if err := envconfig.Process("session", &spec); err != nil {
		envconfig.Usage("session", &spec)
		return
	}
	listener, err := net.Listen("tcp", spec.Addr)
	if err != nil {
		logrus.WithError(err).Fatal("could not setup networking")
	}
	grpcServer := grpc.NewServer()
	api.RegisterSessionServiceServer(grpcServer, session.NewServer(&session.Config{
		Secret:        []byte(spec.Secret),
		RedisAddr:     spec.Redis,
		RedisPassword: spec.RedisPassword,
	}))
	if err := grpcServer.Serve(listener); err != nil {
		logrus.WithError(err).Fatal("failed to serve")
	}
}
