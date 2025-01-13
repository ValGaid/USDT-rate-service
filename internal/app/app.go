package app

import (
	"USDT-rate-service/cmd/config"
	"USDT-rate-service/internal/controller"
	pb "USDT-rate-service/internal/gen/rates"
	"USDT-rate-service/internal/healthcheck"
	"USDT-rate-service/internal/infrastructure/garantexapi"
	"USDT-rate-service/internal/service"
	"USDT-rate-service/internal/storege"
	"USDT-rate-service/pkg/prom"
	"USDT-rate-service/pkg/tracer"
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	maxAttempts    = 5
	waitTime       = 2 * time.Second
	postgresDriver = "postgres"
	protocol       = "tcp"
)

type App struct {
	controller *controller.Controller
	db         *storege.CryptoDB
	metrics    *prom.Metrics
	tracer     trace.Tracer
	stop       chan os.Signal
	address    string
	log        *zap.Logger
}

func NewApp(cfg *config.Config, log *zap.Logger) (*App, error) {
	log.Debug("load config success: ", zap.Any("config:", cfg))
	if err := migrate(cfg, log); err != nil {
		return nil, err
	}

	jaegerURL := fmt.Sprintf("http://%s:%s/api/traces", cfg.TraceHost, cfg.TracePort)
	trc, err := tracer.InitTracer(jaegerURL, log.Name())
	if err != nil {
		return nil, fmt.Errorf("tracer init err: %w", err)
	}

	metrics := prom.NewMetrics()

	pgrDB, err := newPostgresDB(config.GetDataSourceName(cfg), log)
	if err != nil {
		return nil, err
	}

	db := storege.NewUSDTdb(pgrDB, metrics.DBTime, trc)

	client := garantexapi.NewRates(metrics, trc, log)

	serv := service.NewService(client, db, log)

	return &App{
		controller: controller.NewController(serv, metrics),
		db:         db,
		metrics:    metrics,
		tracer:     trc,
		stop:       make(chan os.Signal),
		address:    cfg.Host + ":" + cfg.Port,
		log:        log,
	}, nil
}

func (a *App) Run() {

	ratesServer := grpc.NewServer()

	check := healthcheck.NewCheck(a.db)

	healthgrpc.RegisterHealthServer(ratesServer, check)
	pb.RegisterRatesServer(ratesServer, a.controller)
	listener, err := net.Listen(protocol, a.address)
	if err != nil {
		a.log.Fatal("Failed to starting USDT_rate_service:", zap.Error(err))
	}
	defer listener.Close()

	a.log.Info("USDT_rate_service is listening on", zap.String("address", a.address))

	signal.Notify(a.stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := ratesServer.Serve(listener); err != nil {
			a.log.Fatal("Failed to start USDT_rate_service:", zap.Error(err))
		}
	}()

	a.metrics.InitMetrics()

	<-a.stop

	a.log.Info("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()

	ratesServer.GracefulStop()

	select {
	case <-ctx.Done():
		a.log.Info("Server gracefully shutdown")
	}

	a.log.Info("USDT_rate_service stopped gracefully")
}

func migrate(cfg *config.Config, log *zap.Logger) error {

	log.Debug("starting NewUSDTdb", zap.String("dataSourceName", config.GetDataSourceName(cfg)))

	var db *sqlx.DB

	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = sqlx.Open(postgresDriver, config.GetDataSourceName(cfg))
		if err != nil {
			log.Warn("migrate: failed to open database connection", zap.Error(err))
			time.Sleep(waitTime)
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Warn("migrate: failed to ping database", zap.Error(err))
			time.Sleep(waitTime)
			continue
		}
		if err = goose.Up(db.DB, "."); err != nil {
			return fmt.Errorf("failed to run migrations: %v", zap.Error(err))
		}

		log.Info("Migrations applied successfully!")
		return nil

	}
	db.Close()
	return fmt.Errorf("failed to connect to database after %d attempts: %w", maxAttempts, err)

}

func newPostgresDB(dataSourceName string, log *zap.Logger) (*sql.DB, error) {
	log.Debug("starting NewUserDB", zap.String("dataSourceName", dataSourceName))
	var db *sql.DB

	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = sql.Open(postgresDriver, dataSourceName)
		if err != nil {
			log.Warn("failed to open database connection", zap.Error(err))
			time.Sleep(waitTime)
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Warn("failed to ping database", zap.Error(err))
			time.Sleep(waitTime)
			continue
		}
		log.Info("connected to the database")
		return db, nil
	}
	db.Close()
	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxAttempts, err)
}
