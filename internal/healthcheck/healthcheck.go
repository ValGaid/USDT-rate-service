package healthcheck

import (
	"context"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type DB interface {
	Healthcheck(ctx context.Context) error
}

type HealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	db DB
}

func NewCheck(db DB) *HealthServer {
	return &HealthServer{
		db: db,
	}
}
func (s *HealthServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	_ = in
	if err := s.db.Healthcheck(ctx); err != nil {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}
