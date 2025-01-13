package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

type Metrics struct {
	RequestTime  *prometheus.HistogramVec
	RequestCount *prometheus.CounterVec
	DBTime       *prometheus.HistogramVec
	APITime      *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	m := Metrics{
		RequestTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "request_duration_seconds",
				Help:    "Duration of GRPC requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		RequestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "request_count",
				Help: "Total number of gRPS requests",
			},
			[]string{"method"},
		),
		DBTime: prometheus.NewHistogramVec(

			prometheus.HistogramOpts{
				Name:    "db_duration_seconds",
				Help:    "Duration of DB requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"}),
		APITime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "api_duration_seconds",
			Help:    "Duration of external API requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
			[]string{"method"}),
	}

	return &m
}

// Handler для метрик
func (m *Metrics) InitMetrics() {
	prometheus.MustRegister(m.RequestTime, m.RequestCount, m.DBTime, m.APITime)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
}
