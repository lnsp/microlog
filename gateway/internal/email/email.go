package email

import (
	"context"
	"github.com/lnsp/microlog/gateway/internal/models"
	"github.com/lnsp/microlog/gateway/pkg/tokens"
	"github.com/lnsp/microlog/mail/api"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Client struct {
	data   *models.DataSource
	service string
}

func NewClient(dataSource *models.DataSource, mailService string) *Client {
	return &Client{
		data:   dataSource,
		service: mailService,
	}
}

func (email *Client) Verify(token string, purpose tokens.EmailPurpose) (string, uint, error) {
	conn, err := grpc.Dial(email.service, grpc.WithInsecure())
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to dial mail service")
	}
	defer conn.Close()
	client := api.NewMailServiceClient(conn)
	req := &api.VerificationRequest{
		Token: token,
	}
	switch purpose {
	case tokens.PurposeReset:
		req.Purpose = api.VerificationRequest_PASSWORD_RESET
	case tokens.PurposeConfirmation:
		req.Purpose = api.VerificationRequest_CONFIRMATION
	}
	resp, err := client.VerifyToken(context.Background(), req)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to verify token")
	}
	return resp.Email, uint(resp.UserID), nil
}

func (email *Client) SendConfirmation(userID uint, emailAddr string) error {
	user, err := email.data.User(userID)
	if err != nil {
		return errors.Wrap(err, "failed to find user")
	}
	conn, err := grpc.Dial(email.service, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to dial mail service")
	}
	defer conn.Close()
	client := api.NewMailServiceClient(conn)
	resp, err := client.SendConfirmation(context.Background(), &api.MailRequest{
		Name: user.Name,
		UserID: uint32(userID),
		Email: emailAddr,
	})
	if err != nil {
		return errors.Wrap(err, "failed to send confirmation")
	}
	return nil
}

func (email *Client) SendPasswordReset(userID uint, emailAddr string) error {
	user, err := email.data.User(userID)
	if err != nil {
		return errors.Wrap(err, "failed to find user")
	}
	conn, err := grpc.Dial(email.service, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to dial mail service")
	}
	defer conn.Close()
	client := api.NewMailServiceClient(conn)
	resp, err := client.SendPasswordReset(context.Background(), &api.MailRequest{
		Name: user.Name,
		UserID: uint32(userID),
		Email: emailAddr,
	})
	if err != nil {
		return errors.Wrap(err, "failed to send password reset")
	}
	return nil
}
