package storege

import (
	"USDT-rate-service/internal/models"
	"github.com/prometheus/client_golang/prometheus"

	"context"
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	traceCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const (
	query = `INSERT INTO usdt (timestamp, ask_price, bid_price) VALUES ($1, $2, $3)`
)

type CryptoDB struct {
	db      *sql.DB
	metrics *prometheus.HistogramVec
	tr      trace.Tracer
}

func NewUSDTdb(db *sql.DB, metrics *prometheus.HistogramVec, trc trace.Tracer) *CryptoDB {
	return &CryptoDB{
		db:      db,
		metrics: metrics,
		tr:      trc,
	}
}

func (c *CryptoDB) AddRate(ctx context.Context, rate *models.Rates) error {
	start := time.Now()
	_, span := c.tr.Start(context.Background(), "AddRate")
	defer span.End()

	loc, _ := time.LoadLocation("Europe/Moscow")

	_, err := c.db.ExecContext(ctx, query, time.Unix(rate.Timestamp, 0).In(loc), rate.AskPrice, rate.BidPrice)
	if err != nil {
		if errors.Is(err, ctx.Err()) {
			span.RecordError(err)
			span.SetStatus(traceCodes.Error, err.Error())
			return status.Errorf(codes.Canceled, "failed to add rates: %v", err)
		}
		span.RecordError(err)
		span.SetStatus(traceCodes.Error, err.Error())

		return status.Errorf(codes.Internal, "failed to add rates: %v", err)
	}

	c.metrics.WithLabelValues("AddRate").Observe(time.Since(start).Seconds())
	return nil
}

func (c *CryptoDB) Healthcheck(ctx context.Context) error {
	err := c.db.PingContext(ctx)
	if err != nil {
		if errors.Is(err, ctx.Err()) {
			return status.Error(codes.DeadlineExceeded, err.Error())
		}
		return err
	}
	return nil
}
