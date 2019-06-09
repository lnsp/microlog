// Package session contains an implementation of the SessionServer.
package session

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	health "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/go-redis/redis"
	"github.com/lnsp/microlog/common/logger"
	"github.com/lnsp/microlog/session/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Config stores configuration context for the Session implementation.
type Config struct {
	Secret         []byte
	RedisAddr      string
	RedisPassword  string
	ExpirationTime time.Duration
}

// Server is an implementation of the SessionServer.
type Server struct {
	secret     []byte
	redis      *redis.Client
	expiration time.Duration
}

var log = logger.New()

// Create initiates a new session with the given role and identity context.
func (s *Server) Create(ctx context.Context, req *api.CreateRequest) (*api.CreateResponse, error) {
	log := log.WithFields(logrus.Fields{
		"identity": req.Id,
		"role":     req.Role,
	})
	token, err := s.GenerateToken(&UserInfo{
		Identity: req.Id,
		Role:     req.Role,
	})
	if err != nil {
		log.WithError(err).Warn("failed to generate token")
		return nil, errors.Wrap(err, "failed to generate token")
	}
	if err := s.redis.Set(token, "active", s.expiration).Err(); err != nil {
		log.WithError(err).Warn("failed to save token")
		return nil, errors.Wrap(err, "failed to save token")
	}
	log.Debug("created session")
	return &api.CreateResponse{
		Token: token,
	}, nil
}

// Verify verifies if a given token is valid.
func (s *Server) Verify(ctx context.Context, req *api.VerifyRequest) (*api.VerifyResponse, error) {
	log := log.WithField("token", req.Token)
	info, err := s.ProofToken(req.Token)
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
	if active := s.redis.Get(req.Token).String(); active == "" {
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

// Delete removes an active session from the session store.
func (s *Server) Delete(ctx context.Context, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	log := log.WithField("token", req.Token)
	info, err := s.ProofToken(req.Token)
	if err != nil {
		log.WithError(err).Debug("failed to verify token")
		return nil, errors.Wrap(err, "failed to verify token")
	}
	log = log.WithFields(logrus.Fields{
		"identity": info.Identity,
		"role":     info.Role,
	})
	if err := s.redis.Del(req.Token).Err(); err != nil {
		log.WithError(err).Warn("failed to delete session")
		return nil, errors.Wrap(err, "failed to delete session")
	}
	log.Debug("deleted session")
	return &api.DeleteResponse{}, nil
}

// Health returns an implementation of the GRPC Health Checking service.
func (s *Server) Health() health.HealthServer {
	return &healthServer{s}
}

type healthServer struct {
	s *Server
}

func (h *healthServer) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	// Since we only run one service, no need to check the targeted service.
	// Only dependency of this service is redis, therefore if redis is up, service should be operational.
	ping := h.s.redis.Ping()
	if ping.Err() != nil {
		return &health.HealthCheckResponse{Status: health.HealthCheckResponse_NOT_SERVING}, nil
	}
	return &health.HealthCheckResponse{Status: health.HealthCheckResponse_SERVING}, nil
}

func (h *healthServer) Watch(req *health.HealthCheckRequest, stream health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch not implemented")
}

// NewServer instantiates a new service instance.
func NewServer(cfg *Config) *Server {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	return &Server{
		secret:     cfg.Secret,
		redis:      redisClient,
		expiration: cfg.ExpirationTime,
	}
}
