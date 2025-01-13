package service

import (
	"USDT-rate-service/internal/models"
	"context"
	"go.uber.org/zap"
	"time"
)

type ClientAPI interface {
	GetRatesAPI(ctx context.Context) (*models.Rates, error)
}

type DB interface {
	AddRate(ctx context.Context, rate *models.Rates) error
}

type Service struct {
	client ClientAPI
	db     DB
	log    *zap.Logger
}

func NewService(client ClientAPI, db DB, log *zap.Logger) *Service {
	return &Service{
		client: client,
		db:     db,
		log:    log,
	}
}

func (s *Service) GetRates(ctx context.Context) (*models.Rates, error) {
	resp, err := s.client.GetRatesAPI(ctx)
	if err != nil {
		return nil, err
	}

	go func() {
		ctxDB, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		err := s.db.AddRate(ctxDB, resp)
		if err != nil {
			s.log.Error("Failed to add rate", zap.Error(err))
		}
	}()
	return resp, nil
}
