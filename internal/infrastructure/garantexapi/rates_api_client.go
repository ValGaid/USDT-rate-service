package garantexapi

import (
	"USDT-rate-service/internal/models"
	"USDT-rate-service/pkg/prom"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	url = "https://garantex.org/api/v2/depth?market=usdtrub"
)

type RatesAPI struct {
	client  http.Client
	metrics *prom.Metrics
	tr      trace.Tracer
	log     *zap.Logger
}

func NewRates(metrics *prom.Metrics, trc trace.Tracer, log *zap.Logger) *RatesAPI {
	return &RatesAPI{
		client:  http.Client{},
		metrics: metrics,
		tr:      trc,
		log:     log,
	}
}

func (r RatesAPI) GetRatesAPI(ctx context.Context) (*models.Rates, error) {
	_, span := r.tr.Start(context.Background(), "GetRatesAPI")
	defer span.End()

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error creating new request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			span.SetStatus(codes.Error, err.Error())
			r.log.Debug("rates api request timed out", zap.String("url", url), zap.Any("ms", time.Since(start).Milliseconds()))
			return nil, fmt.Errorf("error executing request: %w", err)
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	r.metrics.APITime.WithLabelValues("GetRatesAPI").Observe(time.Since(start).Seconds())
	defer resp.Body.Close()

	var data Crypto
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		if errors.Is(err, context.Canceled) {
			span.SetStatus(codes.Error, err.Error())
			r.log.Debug("decoding request timed out", zap.String("url", url), zap.Any("ms", time.Since(start).Milliseconds()))
			return nil, fmt.Errorf("decoding request timed out: %w", err)
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	if len(data.Bids) == 0 || len(data.Asks) == 0 {
		span.SetStatus(codes.Error, errors.New("error decoding request: empty response body").Error())
		return nil, fmt.Errorf("error decoding request: empty response body")
	}

	return &models.Rates{
		Timestamp: data.Timestamp,
		AskPrice:  data.Asks[0].Price,
		BidPrice:  data.Bids[0].Price,
	}, nil
}
