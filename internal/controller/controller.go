package controller

import (
	pb "USDT-rate-service/internal/gen/rates"
	"USDT-rate-service/internal/models"
	"USDT-rate-service/pkg/prom"
	"context"
	"time"
)

type ServiceInterface interface {
	GetRates(ctx context.Context) (*models.Rates, error)
}

type Controller struct {
	service ServiceInterface
	pb.UnimplementedRatesServer
	metrics *prom.Metrics
}

func NewController(service ServiceInterface, metrics *prom.Metrics) *Controller {
	return &Controller{
		service: service,
		metrics: metrics,
	}
}

func (c *Controller) GetRates(ctx context.Context, emp *pb.Empty) (*pb.Response, error) {
	_ = emp
	start := time.Now()

	r, err := c.service.GetRates(ctx)
	if err != nil {
		return nil, err
	}
	c.metrics.RequestTime.WithLabelValues("GetRates").Observe(time.Since(start).Seconds())
	c.metrics.RequestCount.WithLabelValues("GetRates").Inc()

	return &pb.Response{
		Timestamp: r.Timestamp,
		AskPrice:  r.AskPrice,
		BidPrice:  r.BidPrice,
	}, nil

}
