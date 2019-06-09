// Package mail provides a GRPC service for sending transactional emails.
package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dgrijalva/jwt-go"
	"github.com/lnsp/microlog/common/logger"
	"github.com/lnsp/microlog/mail/api"
	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	health "google.golang.org/grpc/health/grpc_health_v1"
)

var log = logger.New()

const (
	resetSubject   = "Reset your email"
	confirmSubject = "Please confirm your email"
)

// Config stores the service configuration.
type Config struct {
	Secret                  []byte
	ConfirmURL, ResetURL    string
	SenderName, SenderEmail string
	APIKey                  string
	TemplateFolder          string
}

// Server is the implementation of the mail gRPC service.
type Server struct {
	secret                          []byte
	sender                          *mail.Email
	mail                            *sendgrid.Client
	forgotTemplate, confirmTemplate *template.Template
	apiKey                          string
	confirmURL, resetURL            string
}

// EmailPurpose defines a transaction email purpose.
type EmailPurpose string

const (
	// EmailConfirmation is an account confimation email.
	EmailConfirmation EmailPurpose = "purpose_confirmation"
	// EmailPasswordReset is a password reset email.
	EmailPasswordReset = "purpose_resetpassword"
)

var (
	// ExpirationTimes defines the expiration times of special-purpose tokens.
	ExpirationTimes = map[EmailPurpose]time.Duration{
		EmailConfirmation:  time.Hour * 72,
		EmailPasswordReset: time.Hour,
	}
	// MapPurpose defines a map of verification request purposes to EmailPurposes.
	MapPurpose = map[api.VerificationRequest_Purpose]EmailPurpose{
		api.VerificationRequest_CONFIRMATION:   EmailConfirmation,
		api.VerificationRequest_PASSWORD_RESET: EmailPasswordReset,
	}
)

// EmailInfo stores general information about the email.
type EmailInfo struct {
	Identity     uint32
	EmailAddress string
	Purpose      EmailPurpose
}

// Claims stores email info in a JWT-compatible way.
type Claims struct {
	jwt.StandardClaims
	EmailInfo
}

type emailContext struct {
	Name, Link string
}

// GenerateToken generates a new token based on the given email information.
func (s *Server) GenerateToken(info *EmailInfo) (string, error) {
	expiration := ExpirationTimes[info.Purpose]
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiration).Unix(),
		},
		EmailInfo: *info,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign claims")
	}
	return signed, nil
}

// ProofToken checks if the given token string is valid.
func (s *Server) ProofToken(signed string) (*EmailInfo, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(signed, &claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token")
	}
	return &claims.EmailInfo, nil
}

// VerifyToken verifies the given special-purpose token.
func (s *Server) VerifyToken(ctx context.Context, req *api.VerificationRequest) (*api.VerificationResponse, error) {
	log := log.WithFields(logrus.Fields{
		"purpose": req.Purpose,
		"token":   req.Token,
	})
	info, err := s.ProofToken(req.Token)
	if err != nil {
		log.WithError(err).Warn("verification failed")
		return nil, errors.Wrap(err, "verification failed")
	}
	if info.Purpose != MapPurpose[req.Purpose] {
		log.Warn("purpose does not match")
		return nil, errors.New("purpose does not match")
	}
	log.WithFields(logrus.Fields{
		"identity": info.Identity,
		"email":    info.EmailAddress,
	}).Debug("verification successful")
	return &api.VerificationResponse{
		Email: info.EmailAddress,
		Id:    info.Identity,
	}, nil
}

// SendConfirmation sends a account confirmation email.
func (s *Server) SendConfirmation(ctx context.Context, req *api.MailRequest) (*api.MailResponse, error) {
	log := log.WithFields(logrus.Fields{
		"email":    req.Email,
		"identity": req.Id,
		"purpose":  EmailConfirmation,
	})
	token, err := s.GenerateToken(&EmailInfo{
		EmailAddress: req.Email,
		Identity:     req.Id,
		Purpose:      EmailConfirmation,
	})
	if err != nil {
		log.WithError(err).Warn("failed to create token")
		return nil, errors.Wrap(err, "failed to create token")
	}
	link := fmt.Sprintf(s.confirmURL, token)
	buf := new(bytes.Buffer)
	if err := s.confirmTemplate.Execute(buf, &emailContext{Name: req.Name, Link: link}); err != nil {
		log.WithError(err).Warn("failed to render email")
		return nil, errors.Wrap(err, "failed to render email")
	}
	receiver := mail.NewEmail(req.Name, req.Email)
	message := mail.NewSingleEmail(s.sender, confirmSubject, receiver, buf.String(), buf.String())
	resp, err := s.mail.Send(message)
	if err != nil {
		log.WithError(err).Warn("failed to send email")
		return nil, errors.Wrap(err, "failed to send email")
	}
	log.WithFields(logrus.Fields{
		"link":   link,
		"status": resp.StatusCode,
	}).Debug("sent confirmation email")
	return &api.MailResponse{
		Status: "OK",
		Code:   int32(resp.StatusCode),
	}, nil
}

// SendPasswordReset sends a password reset email.
func (s *Server) SendPasswordReset(ctx context.Context, req *api.MailRequest) (*api.MailResponse, error) {
	log := log.WithFields(logrus.Fields{
		"email":    req.Email,
		"identity": req.Id,
		"purpose":  EmailPasswordReset,
	})
	token, err := s.GenerateToken(&EmailInfo{
		EmailAddress: req.Email,
		Identity:     req.Id,
		Purpose:      EmailPasswordReset,
	})
	if err != nil {
		log.WithError(err).Warn("failed to create token")
		return nil, errors.Wrap(err, "failed to create token")
	}
	link := fmt.Sprintf(s.resetURL, token)
	buf := new(bytes.Buffer)
	if err := s.forgotTemplate.Execute(buf, &emailContext{Name: req.Name, Link: link}); err != nil {
		log.WithError(err).Warn("failed to render email")
		return nil, errors.Wrap(err, "failed to render email")
	}
	receiver := mail.NewEmail(req.Name, req.Email)
	message := mail.NewSingleEmail(s.sender, resetSubject, receiver, buf.String(), buf.String())
	resp, err := s.mail.Send(message)
	if err != nil {
		log.WithError(err).Warn("failed to send email")
		return nil, errors.Wrap(err, "failed to send email")
	}
	log.WithFields(logrus.Fields{
		"link":   link,
		"status": resp.StatusCode,
	}).Debug("sent password reset email")
	return &api.MailResponse{
		Status: "OK",
		Code:   int32(resp.StatusCode),
	}, nil
}

// HealthServer provides an implementation of the GRPC Health Checking Protocol.
func (s *Server) HealthServer() health.HealthServer {
	return &healthServer{s}
}

type healthServer struct {
	s *Server
}

func (h *healthServer) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	// We will ignore the service parameter for now since we only implement one service per package.
	// Check for sendgrid if everything works since mail is important
	_, err := sendgrid.MakeRequest(sendgrid.GetRequest(h.s.apiKey, "/api/v3/alerts", ""))
	if err != nil {
		return &health.HealthCheckResponse{Status: health.HealthCheckResponse_NOT_SERVING}, nil
	}
	return &health.HealthCheckResponse{Status: health.HealthCheckResponse_SERVING}, nil
}

func (h *healthServer) Watch(req *health.HealthCheckRequest, stream health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch is not implemented")
}

// NewServer sets up a new gRPC server instance.
func NewServer(cfg *Config) *Server {
	forgotTemplate := template.Must(template.ParseFiles(cfg.TemplateFolder + "/forgot.html"))
	confirmTemplate := template.Must(template.ParseFiles(cfg.TemplateFolder + "/confirm.html"))
	return &Server{
		sender:          mail.NewEmail(cfg.SenderName, cfg.SenderEmail),
		mail:            sendgrid.NewSendClient(cfg.APIKey),
		apiKey:          cfg.APIKey,
		forgotTemplate:  forgotTemplate,
		confirmTemplate: confirmTemplate,
		resetURL:        cfg.ResetURL,
		confirmURL:      cfg.ConfirmURL,
	}
}
