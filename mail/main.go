package main

import (
	"net"

	health "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/kelseyhightower/envconfig"
	"github.com/lnsp/microlog/common/logger"
	"github.com/lnsp/microlog/mail/api"
	"github.com/lnsp/microlog/mail/internal/mail"
	"google.golang.org/grpc"
)

type specification struct {
	APIKey      string `required:"true" desc:"SendGrid API Key"`
	Secret      string `required:"true" desc:"Encryption secret for tokens"`
	Addr        string `default:":8080" desc:"Host and port to listen on"`
	ConfirmURL  string `default:"http://localhost:8080/auth/confirm?token=%s" desc:"Confirmation URL format"`
	ResetURL    string `default:"http://localhost:8080/auth/reset?token=%s" desc:"Reset URL format"`
	Templates   string `default:"templates" desc:"Template folder"`
	SenderName  string `default:"The microlog team" desc:"The default sender name"`
	SenderEmail string `default:"team@microlog.co" desc:"The default sender email"`
}

var log = logger.New()

func main() {
	var spec specification
	if err := envconfig.Process("mail", &spec); err != nil {
		envconfig.Usage("mail", &spec)
		return
	}
	listener, err := net.Listen("tcp", spec.Addr)
	if err != nil {
		log.WithError(err).Fatal("could not setup networking")
	}
	grpcServer := grpc.NewServer()
	mailServer := mail.NewServer(&mail.Config{
		APIKey:         spec.APIKey,
		TemplateFolder: spec.Templates,
		ConfirmURL:     spec.ConfirmURL,
		ResetURL:       spec.ResetURL,
		SenderName:     spec.SenderName,
		SenderEmail:    spec.SenderEmail,
		Secret:         []byte(spec.Secret),
	})
	api.RegisterMailServer(grpcServer, mailServer)
	health.RegisterHealthServer(grpcServer, mailServer.Health())
	if err := grpcServer.Serve(listener); err != nil {
		log.WithError(err).Fatal("could not serve")
	}
}
