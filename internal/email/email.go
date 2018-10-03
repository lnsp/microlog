package email

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/sendgrid/sendgrid-go"

	"github.com/alecthomas/template"
	"github.com/lnsp/microlog/internal/models"
	"github.com/lnsp/microlog/internal/tokens"
	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	confirmSubject = "Please confirm your email"
	resetSubject   = "Reset your password"
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.JSONFormatter),
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.DebugLevel,
}

var (
	noreplySender   = mail.NewEmail("The microlog team", "noreply@microlog.co")
	confirmTemplate = template.Must(template.ParseFiles("./web/templates/email/confirm.html"))
	forgotTemplate  = template.Must(template.ParseFiles("./web/templates/email/forgot.html"))
)

type Client struct {
	client *sendgrid.Client
	data   *models.DataSource
	secret []byte
}

func NewClient(dataSource *models.DataSource, tokenSecret []byte, apiKey string) *Client {
	return &Client{
		data:   dataSource,
		secret: tokenSecret,
		client: sendgrid.NewSendClient(apiKey),
	}
}

type emailContext struct {
	Name, Link string
}

func (email *Client) SendConfirmation(userID uint, emailAddr string, url string) error {
	user, err := email.data.User(userID)
	if err != nil {
		return errors.Wrap(err, "failed to find user")
	}
	token, err := tokens.CreateEmailToken(email.secret, emailAddr, userID, tokens.PurposeConfirmation)
	if err != nil {
		return errors.Wrap(err, "failed to create email token")
	}
	link := fmt.Sprintf(url, token)
	buf := new(bytes.Buffer)
	if err := confirmTemplate.Execute(buf, &emailContext{Name: user.Name, Link: link}); err != nil {
		return errors.Wrap(err, "failed to render email")
	}
	receiver := mail.NewEmail(user.Name, emailAddr)
	message := mail.NewSingleEmail(noreplySender, confirmSubject, receiver, buf.String(), buf.String())
	resp, err := email.client.Send(message)
	if err != nil {
		return errors.Wrap(err, "failed to send email")
	}
	log.WithFields(logrus.Fields{
		"type":   "confirmation",
		"addr":   emailAddr,
		"token":  token,
		"status": resp.StatusCode,
	}).Debug("send confirmation email")
	return nil
}

func (email *Client) SendPasswordReset(userID uint, emailAddr, url string) error {
	user, err := email.data.User(userID)
	if err != nil {
		return errors.Wrap(err, "failed to find user")
	}
	token, err := tokens.CreateEmailToken(email.secret, emailAddr, userID, tokens.PurposeReset)
	if err != nil {
		return errors.Wrap(err, "failed to create token")
	}
	link := fmt.Sprintf(url, token)
	buf := new(bytes.Buffer)
	if err := forgotTemplate.Execute(buf, &emailContext{Name: user.Name, Link: link}); err != nil {
		return errors.Wrap(err, "failed to render email")
	}
	receiver := mail.NewEmail(user.Name, emailAddr)
	message := mail.NewSingleEmail(noreplySender, resetSubject, receiver, buf.String(), buf.String())
	resp, err := email.client.Send(message)
	if err != nil {
		return errors.Wrap(err, "failed to send email")
	}
	log.WithFields(logrus.Fields{
		"type":   "passwordReset",
		"addr":   emailAddr,
		"token":  token,
		"status": resp.StatusCode,
	}).Debug("send password reset email")
	return nil
}
