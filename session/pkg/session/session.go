package session

import (
	"github.com/go-redis/redis"
	"github.com/lnsp/microlog/session/api"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"time"
)

type Config struct {
	Secret         []byte
	RedisAddr      string
	RedisPassword  string
	ExpirationTime time.Duration
}

type Server struct {
	secret     []byte
	redis      *redis.Client
	expiration time.Duration
}

func (svc *Server) Create(ctx context.Context, req *api.CreateRequest) (*api.CreateResponse, error) {
	token, err := svc.GenerateToken(&UserInfo{
		Identity: req.Id,
		Role:     req.Role,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate token")
	}
	if err := svc.redis.Set(token, "active", svc.expiration).Err(); err != nil {
		return nil, errors.Wrap(err, "failed to save toen")
	}
	return &api.CreateResponse{
		Token: token,
	}, nil
}

func (svc *Server) Verify(ctx context.Context, req *api.VerifyRequest) (*api.VerifyResponse, error) {
	info, err := svc.ProofToken(req.Token)
	if err != nil {
		return &api.VerifyResponse{
			Ok: false,
		}, nil
	}
	if active := svc.redis.Get(req.Token).String(); active != "active" {
		return &api.VerifyResponse{
			Ok: false,
		}, nil
	}
	return &api.VerifyResponse{
		Ok:   true,
		Id:   info.Identity,
		Role: info.Role,
	}, nil
}

func (svc *Server) Delete(ctx context.Context, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	_, err := svc.ProofToken(req.Token)
	if err != nil {
		return nil, errors.Wrap(err, "could not validate token")
	}
	if err := svc.redis.Del(req.Token).Err(); err != nil {
		return nil, errors.Wrap(err, "could not delete session")
	}
	return &api.DeleteResponse{}, nil
}

func NewServer(cfg *Config) *Server {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	return &Server{
		secret: cfg.Secret,
		redis:  redisClient,
	}
}
