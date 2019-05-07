package email

import (
	"context"
	"time"

	"github.com/lnsp/microlog/gateway/internal/models"
	"github.com/lnsp/microlog/mail/api"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Client struct {
	data    *models.DataSource
	service string
}

func NewClient(dataSource *models.DataSource, mailService string) *Client {
	return &Client{
		data:    dataSource,
		service: mailService,
	}
}

func (email *Client) serviceClient() (api.MailServiceClient, *grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, email.service, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to dial mail service %s", email.service)
	}
	client := api.NewMailServiceClient(conn)
	return client, conn, nil
}

func (email *Client) VerifyConfirmationToken(token string) (string, uint, error) {
	return email.verifyToken(token, api.VerificationRequest_CONFIRMATION)
}

func (email *Client) VerifyPasswordResetToken(token string) (string, uint, error) {
	return email.verifyToken(token, api.VerificationRequest_PASSWORD_RESET)
}

func (email *Client) verifyToken(token string, purpose api.VerificationRequest_Purpose) (string, uint, error) {
	client, conn, err := email.serviceClient()
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to create client")
	}
	defer conn.Close()
	req := &api.VerificationRequest{
		Token:   token,
		Purpose: purpose,
	}
	resp, err := client.VerifyToken(context.Background(), req)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to verify token")
	}
	return resp.Email, uint(resp.Id), nil
}

func (email *Client) SendConfirmation(userID uint, emailAddr string) error {
	user, err := email.data.User(userID)
	if err != nil {
		return errors.Wrap(err, "failed to find user")
	}
	client, conn, err := email.serviceClient()
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}
	defer conn.Close()
	resp, err := client.SendConfirmation(context.Background(), &api.MailRequest{
		Name:  user.Name,
		Id:    uint32(userID),
		Email: emailAddr,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to send confirmation: %v", resp)
	}
	return nil
}

func (email *Client) SendPasswordReset(userID uint, emailAddr string) error {
	user, err := email.data.User(userID)
	if err != nil {
		return errors.Wrap(err, "failed to find user")
	}
	client, conn, err := email.serviceClient()
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}
	defer conn.Close()
	resp, err := client.SendPasswordReset(context.Background(), &api.MailRequest{
		Name:  user.Name,
		Id:    uint32(userID),
		Email: emailAddr,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to send password reset: %v", resp)
	}
	return nil
}
