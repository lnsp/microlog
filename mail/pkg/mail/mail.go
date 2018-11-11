package mail

import (
	"bytes"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/lnsp/microlog/mail/api"
	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"html/template"
	"time"
)

const (
	resetSubject   = "Reset your email"
	confirmSubject = "Please confirm your email"
)

type Config struct {
	Secret                  []byte
	ConfirmURL, ResetURL    string
	SenderName, SenderEmail string
	APIKey                  string
	TemplateFolder          string
}

type Server struct {
	secret                          []byte
	sender                          *mail.Email
	mail                            *sendgrid.Client
	forgotTemplate, confirmTemplate *template.Template
	apiKey                          string
	confirmURL, resetURL            string
}

type EmailPurpose string

const (
	EmailConfirmation  EmailPurpose = "purpose_confirmation"
	EmailPasswordReset              = "purpose_resetpassword"
)

var (
	ExpirationTimes = map[EmailPurpose]time.Duration{
		EmailConfirmation:  time.Hour * 72,
		EmailPasswordReset: time.Hour,
	}
	MapPurpose = map[api.VerificationRequest_Purpose]EmailPurpose{
		api.VerificationRequest_CONFIRMATION:   EmailConfirmation,
		api.VerificationRequest_PASSWORD_RESET: EmailPasswordReset,
	}
)

type EmailInfo struct {
	Identity     uint32
	EmailAddress string
	Purpose      EmailPurpose
}

type Claims struct {
	jwt.StandardClaims
	EmailInfo
}

type emailContext struct {
	Name, Link string
}

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

func (s *Server) VerifyToken(ctx context.Context, req *api.VerificationRequest) (*api.VerificationResponse, error) {
	info, err := s.ProofToken(req.Token)
	if err != nil {
		return nil, errors.Wrap(err, "verification failed")
	}
	if info.Purpose != MapPurpose[req.Purpose] {
		return nil, errors.New("purpose does not match")
	}
	return &api.VerificationResponse{
		Email: info.EmailAddress,
		Id:    info.Identity,
	}, nil
}

func (s *Server) SendConfirmation(ctx context.Context, req *api.MailRequest) (*api.MailResponse, error) {
	token, err := s.GenerateToken(&EmailInfo{
		EmailAddress: req.Email,
		Identity:     req.Id,
		Purpose:      EmailConfirmation,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create token")
	}
	link := fmt.Sprintf(s.confirmURL, token)
	buf := new(bytes.Buffer)
	if err := s.confirmTemplate.Execute(buf, &emailContext{Name: req.Name, Link: link}); err != nil {
		return nil, errors.Wrap(err, "failed to render email")
	}
	receiver := mail.NewEmail(req.Name, req.Email)
	message := mail.NewSingleEmail(s.sender, confirmSubject, receiver, buf.String(), buf.String())
	resp, err := s.mail.Send(message)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send email")
	}
	logrus.WithFields(logrus.Fields{
		"type":   "passwordReset",
		"addr":   req.Email,
		"link":   link,
		"status": resp.StatusCode,
	}).Debug("send password reset email")
	return &api.MailResponse{
		Status: "OK",
		Code:   int32(resp.StatusCode),
	}, nil
}

func (s *Server) SendPasswordReset(ctx context.Context, req *api.MailRequest) (*api.MailResponse, error) {
	token, err := s.GenerateToken(&EmailInfo{
		EmailAddress: req.Email,
		Identity:     req.Id,
		Purpose:      EmailConfirmation,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create token")
	}
	link := fmt.Sprintf(s.resetURL, token)
	buf := new(bytes.Buffer)
	if err := s.forgotTemplate.Execute(buf, &emailContext{Name: req.Name, Link: link}); err != nil {
		return nil, errors.Wrap(err, "failed to render email")
	}
	receiver := mail.NewEmail(req.Name, req.Email)
	message := mail.NewSingleEmail(s.sender, resetSubject, receiver, buf.String(), buf.String())
	resp, err := s.mail.Send(message)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send email")
	}
	logrus.WithFields(logrus.Fields{
		"type":   "passwordReset",
		"addr":   req.Email,
		"link":   link,
		"status": resp.StatusCode,
	}).Debug("send password reset email")
	return &api.MailResponse{
		Status: "OK",
		Code:   int32(resp.StatusCode),
	}, nil
}

func NewServer(cfg *Config) *Server {
	forgotTemplate := template.Must(template.ParseFiles(cfg.TemplateFolder + "/forgot.html"))
	confirmTemplate := template.Must(template.ParseFiles(cfg.TemplateFolder + "/confirm.html"))
	return &Server{
		sender:          mail.NewEmail(cfg.SenderName, cfg.SenderEmail),
		mail:            sendgrid.NewSendClient(cfg.APIKey),
		forgotTemplate:  forgotTemplate,
		confirmTemplate: confirmTemplate,
		resetURL:        cfg.ResetURL,
		confirmURL:      cfg.ConfirmURL,
	}
}
