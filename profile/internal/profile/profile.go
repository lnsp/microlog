package profile

import (
	"context"

	"github.com/lnsp/microlog/common/logger"
	"github.com/lnsp/microlog/profile/api"
	"github.com/lnsp/microlog/profile/internal/profile/models"
	"github.com/minio/minio-go"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var log = logger.New()

type Config struct {
	Datasource   string
	S3AccessKey  string
	S3SecretKey  string
	S3BucketPath string
	S3Bucket     string
	S3Endpoint   string
}

func NewServer(cfg *Config) api.ProfileServer {
	client, err := minio.New(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, true)
	if err != nil {
		log.WithError(err).Fatal("could not create minio client")
	}
	data, err := models.Open(cfg.Datasource)
	if err != nil {
		log.WithError(err).Fatal("could not connect to data source")
	}
	return &ProfileServer{
		s3: client,
		db: data,
	}
}

type ProfileServer struct {
	s3 *minio.Client
	db *models.DB
}

func (s *ProfileServer) Create(ctx context.Context, req *api.ProfileCreateRequest) (*api.ProfileCreateResponse, error) {
	return nil, status.Error(codes.Unimplemented, "create not implemented")
}

func (s *ProfileServer) Delete(ctx context.Context, req *api.ProfileDeleteRequest) (*api.ProfileDeleteResponse, error) {
	return nil, status.Error(codes.Unimplemented, "delete not implemented")
}

func (s *ProfileServer) Get(ctx context.Context, req *api.ProfileGetRequest) (*api.ProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "get not implemented")
}

func (s *ProfileServer) UpdateBiography(ctx context.Context, req *api.ProfileUpdateRequest) (*api.ProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "update_biography not implemented")
}

func (s *ProfileServer) UpdateDisplayName(ctx context.Context, req *api.ProfileUpdateRequest) (*api.ProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "update_displayname not implemented")
}

func (s *ProfileServer) UpdateImage(ctx context.Context, req *api.ProfileUpdateRequest) (*api.ProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "update_image not implemented")
}

// Health returns an implementation of the GRPC Health Checking protocol.
func (s *ProfileServer) Health() health.HealthServer {
	return &healthServer{s}
}

type healthServer struct {
	s *ProfileServer
}

func (h *healthServer) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	// Since we have only one service running, we don not need to check the health target string.
	// First check if S3-compatible bucket is accessible.
	if _, err := h.s.s3.ListBuckets(); err != nil {
		log.WithError(err).Error("S3-compatible storage unreachable")
		return &health.HealthCheckResponse{
			Status: health.HealthCheckResponse_NOT_SERVING,
		}, nil
	}
	// After that check for db.
	if err := h.s.db.Ping(); err != nil {
		log.WithError(err).Error("postgresql storage unreachable")
		return &health.HealthCheckResponse{
			Status: health.HealthCheckResponse_NOT_SERVING,
		}, nil
	}
	// Should be fine then.
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

func (h *healthServer) Watch(req *health.HealthCheckRequest, stream health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch not implemented")
}
