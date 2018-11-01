package mail

import (
	"bytes"
	"fmt"
	"github.com/lnsp/microlog/gateway/pkg/tokens"
	"github.com/lnsp/microlog/mail/api"
	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"html/template"
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

type MailServer struct {
	secret                          []byte
	sender                          *mail.Email
	mail                            *sendgrid.Client
	forgotTemplate, confirmTemplate *template.Template
	apiKey                          string
	confirmURL, resetURL            string
}

type emailContext struct {
	Name, Link string
}

func (s *MailServer) VerifyToken(ctx context.Context, req *api.VerificationRequest) (*api.VerificationResponse, error) {
	var purpose tokens.EmailPurpose
	switch req.Purpose {
	case api.VerificationRequest_CONFIRMATION:
		purpose = tokens.PurposeConfirmation
	case api.VerificationRequest_PASSWORD_RESET:
		purpose = tokens.PurposeReset
	}
	email, userID, ok := tokens.VerifyEmailToken(s.secret, req.Token, purpose)
	if !ok {
		return nil, errors.New("verification failed")
	}
	return &api.VerificationResponse{
		Email: email,
		UserID: uint32(userID),
	}, nil
}

func (s *MailServer) SendConfirmation(ctx context.Context, req *api.MailRequest) (*api.MailResponse, error) {
	token, err := tokens.CreateEmailToken(s.secret, req.Email, uint(req.UserID), tokens.PurposeConfirmation)
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

func (s *MailServer) SendPasswordReset(ctx context.Context, req *api.MailRequest) (*api.MailResponse, error) {
	token, err := tokens.CreateEmailToken(s.secret, req.Email, uint(req.UserID), tokens.PurposeReset)
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

func NewServer(cfg *Config) *MailServer {
	forgotTemplate := template.Must(template.ParseFiles(cfg.TemplateFolder + "/forgot.html"))
	confirmTemplate := template.Must(template.ParseFiles(cfg.TemplateFolder + "/confirm.html"))
	return &MailServer{
		sender:          mail.NewEmail(cfg.SenderName, cfg.SenderEmail),
		mail:            sendgrid.NewSendClient(cfg.APIKey),
		forgotTemplate:  forgotTemplate,
		confirmTemplate: confirmTemplate,
		resetURL:        cfg.ResetURL,
		confirmURL:      cfg.ConfirmURL,
	}
}
