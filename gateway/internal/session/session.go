package session

import (
	"github.com/lnsp/microlog/gateway/internal/models"
	"github.com/lnsp/microlog/session/api"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"time"
)

type Client struct {
	data    *models.DataSource
	service string
}

func NewClient(dataSource *models.DataSource, sessionService string) *Client {
	return &Client{
		data:    dataSource,
		service: sessionService,
	}
}

func (session *Client) serviceClient() (api.SessionServiceClient, *grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, session.service, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to dial session service")
	}
	service := api.NewSessionServiceClient(conn)
	return service, conn, nil
}

func (session *Client) Create(userID uint) (string, error) {
	user, err := session.data.User(userID)
	if err != nil {
		return "", errors.Wrap(err, "could not create context")
	}
	client, conn, err := session.serviceClient()
	if err != nil {
		return "", errors.Wrap(err, "failed to create client")
	}
	defer conn.Close()
	role := "user"
	if user.Moderator {
		role = "moderator"
	}
	resp, err := client.Create(context.Background(), &api.CreateRequest{
		Id:   uint32(user.ID),
		Role: role,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to create token")
	}
	return resp.Token, nil
}

func (session *Client) Verify(token string) (uint, bool, error) {
	client, conn, err := session.serviceClient()
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to create client")
	}
	defer conn.Close()
	resp, err := client.Verify(context.Background(), &api.VerifyRequest{
		Token: token,
	})
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to verify token")
	}
	if !resp.Ok {
		return 0, false, errors.Wrap(err, "token not accepted")
	}
	return uint(resp.Id), resp.Role == "moderator", nil
}

func (session *Client) Delete(token string) error {
	client, conn, err := session.serviceClient()
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}
	defer conn.Close()
	_, err = client.Delete(context.Background(), &api.DeleteRequest{
		Token: token,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete token")
	}
	return nil
}
