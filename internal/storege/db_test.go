package storege

import (
	"USDT-rate-service/internal/models"
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"log"
	"testing"
	"time"
)

func getMockTimescaleDB() (*CryptoDB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		log.Fatalf("failed to open a stub database connection: %s", err)
	}
	var mockMetric = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_duration_seconds",
			Help:    "Duration of DB requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"})

	return &CryptoDB{
		db:      db,
		metrics: mockMetric,
		tr:      trace.NewNoopTracerProvider().Tracer("test"),
	}, mock

}
func TestAddRate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	tests := []struct {
		name           string
		err            error
		ctx            context.Context
		rate           *models.Rates
		expectedResult int64
	}{
		{
			name: "success",
			rate: &models.Rates{
				Timestamp: 1736348686,
				AskPrice:  "105.16",
				BidPrice:  "105.14",
			},
			err:            nil,
			ctx:            context.Background(),
			expectedResult: 1,
		},
		{
			name: "Canceled",
			rate: &models.Rates{
				Timestamp: 1736348686,
				AskPrice:  "105.16",
				BidPrice:  "105.14",
			},
			err:            fmt.Errorf("rpc error: code = Canceled"),
			ctx:            ctx,
			expectedResult: 0,
		},
		{
			name: "error",
			rate: &models.Rates{
				Timestamp: 1736348686,
				AskPrice:  "",
				BidPrice:  "105.14",
			},
			err:            fmt.Errorf("rpc error: code = Internal"),
			ctx:            context.Background(),
			expectedResult: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, mock := getMockTimescaleDB()
			loc, _ := time.LoadLocation("Europe/Moscow")
			if tt.name == "success" {

				mock.ExpectExec("INSERT INTO usdt").
					WithArgs(time.Unix(tt.rate.Timestamp, 0).In(loc), tt.rate.AskPrice, tt.rate.BidPrice).
					WillReturnResult(sqlmock.NewResult(0, 1))
			}
			if tt.name == "Canceled" {
				mock.ExpectExec("INSERT INTO usdt").
					WithArgs(time.Unix(tt.rate.Timestamp, 0).In(loc), tt.rate.AskPrice, tt.rate.BidPrice)
			} else {
				mock.ExpectExec("INSERT INTO usdt").
					WithArgs(time.Unix(tt.rate.Timestamp, 0).In(loc), tt.rate.AskPrice, tt.rate.BidPrice).
					WillReturnError(fmt.Errorf("failed to insert rate"))
			}

			err := storage.AddRate(tt.ctx, tt.rate)
			if tt.err != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.err.Error())
				assert.ObjectsAreEqual(t, tt.expectedResult) // Проверяем, что ошибка содержит ожидаемое сообщение
			} else {
				assert.NoError(t, err)
				assert.ObjectsAreEqual(t, tt.expectedResult)
			}
		})

	}
}
func TestHealthcheck(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Nanosecond)
	defer cancel()
	tests := []struct {
		name          string
		ctx           context.Context
		mockPingError error
	}{
		{
			name:          "success",
			ctx:           context.Background(),
			mockPingError: nil,
		},
		{
			name:          "ping error",
			ctx:           context.Background(),
			mockPingError: fmt.Errorf("ping error"),
		},
		{
			name:          "context canceled",
			ctx:           ctx,
			mockPingError: fmt.Errorf("rpc error: code = DeadlineExceeded"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, mock := getMockTimescaleDB()

			if tt.mockPingError != nil {
				mock.ExpectPing().WillReturnError(tt.mockPingError)
			} else {
				mock.ExpectPing()
			}

			err := storage.Healthcheck(tt.ctx)
			if tt.mockPingError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.mockPingError.Error())
			} else if tt.mockPingError == nil {
				assert.NoError(t, err)
			}

		})
	}
}
func TestNewUSDTdb(t *testing.T) {

	db := &sql.DB{}
	metrics := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "test_histogram",
		Help: "A test histogram",
	}, []string{"label"})
	trc := trace.NewNoopTracerProvider().Tracer("test")

	cryptoDB := NewUSDTdb(db, metrics, trc)

	if cryptoDB == nil {
		t.Fatal("Expected non-nil CryptoDB")
	}

	if cryptoDB.db != db {
		t.Errorf("Expected db to be %v, got %v", db, cryptoDB.db)
	}
	if cryptoDB.metrics != metrics {
		t.Errorf("Expected metrics to be %v, got %v", metrics, cryptoDB.metrics)
	}
	if cryptoDB.tr != trc {
		t.Errorf("Expected tracer to be %v, got %v", trc, cryptoDB.tr)
	}
}
