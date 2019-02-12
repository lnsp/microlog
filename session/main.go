package main

import (
	"net"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/lnsp/microlog/common"
	"github.com/lnsp/microlog/session/api"
	"github.com/lnsp/microlog/session/pkg/session"
	"google.golang.org/grpc"
)

var log = common.Logger()

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
	grpcServer := grpc.NewServer()
	expirationTime, err := time.ParseDuration(spec.Expiration)
	api.RegisterSessionServiceServer(grpcServer, session.NewServer(&session.Config{
		Secret:         []byte(spec.Secret),
		RedisAddr:      spec.Redis,
		RedisPassword:  spec.RedisPassword,
		ExpirationTime: expirationTime,
	}))
	if err := grpcServer.Serve(listener); err != nil {
		log.WithError(err).Fatal("failed to serve")
	}
}
