package session

import (
	"github.com/go-redis/redis"
	"github.com/lnsp/microlog/common"
	"github.com/lnsp/microlog/session/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

var log = common.Logger()

func (svc *Server) Create(ctx context.Context, req *api.CreateRequest) (*api.CreateResponse, error) {
	log := log.WithFields(logrus.Fields{
		"identity": req.Id,
		"role":     req.Role,
	})
	token, err := svc.GenerateToken(&UserInfo{
		Identity: req.Id,
		Role:     req.Role,
	})
	if err != nil {
		log.WithError(err).Warn("failed to generate token")
		return nil, errors.Wrap(err, "failed to generate token")
	}
	if err := svc.redis.Set(token, "active", svc.expiration).Err(); err != nil {
		log.WithError(err).Warn("failed to save token")
		return nil, errors.Wrap(err, "failed to save token")
	}
	log.Debug("created session")
	return &api.CreateResponse{
		Token: token,
	}, nil
}

func (svc *Server) Verify(ctx context.Context, req *api.VerifyRequest) (*api.VerifyResponse, error) {
	log := log.WithField("token", req.Token)
	info, err := svc.ProofToken(req.Token)
	if err != nil {
		log.WithError(err).Debug("failed to verify token")
		return &api.VerifyResponse{
			Ok: false,
		}, nil
	}
	log = log.WithFields(logrus.Fields{
		"identity": info.Identity,
		"role":     info.Role,
	})
	if active := svc.redis.Get(req.Token).String(); active != "" {
		log.WithError(err).Warn("attempt to sign in using deleted session")
		return &api.VerifyResponse{
			Ok: false,
		}, nil
	}
	log.Debug("verified session")
	return &api.VerifyResponse{
		Ok:   true,
		Id:   info.Identity,
		Role: info.Role,
	}, nil
}

func (svc *Server) Delete(ctx context.Context, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	log := log.WithField("token", req.Token)
	info, err := svc.ProofToken(req.Token)
	if err != nil {
		log.WithError(err).Debug("failed to verify token")
		return nil, errors.Wrap(err, "failed to verify token")
	}
	log = log.WithFields(logrus.Fields{
		"identity": info.Identity,
		"role":     info.Role,
	})
	if err := svc.redis.Del(req.Token).Err(); err != nil {
		log.WithError(err).Warn("failed to delete session")
		return nil, errors.Wrap(err, "failed to delete session")
	}
	log.Debug("deleted session")
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
