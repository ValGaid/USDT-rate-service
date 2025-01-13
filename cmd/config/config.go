package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Host     string `env:"HOST"`
	Port     string `env:"PORT"`
	LogLevel string `env:"LOG_LEVEL"`

	TraceHost string `env:"TRACE_HOST"`
	TracePort string `env:"TRACE_PORT"`

	DBHost   string `env:"DB_HOST"`
	DBPort   string `env:"DB_PORT"`
	DBUser   string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD" `
	DBName   string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSLMODE"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Host, "host", cfg.Host, "Server host (e.g. localhost)")
	flag.StringVar(&cfg.Port, "port", cfg.Port, "Server port (e.g. 50051)")

	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Logging level (e.g. DEBUG, INFO, WARN, ERROR)")

	flag.StringVar(&cfg.TraceHost, "trace-host", cfg.TraceHost, "Trace service host (e.g. local)")
	flag.StringVar(&cfg.TracePort, "trace-port", cfg.TracePort, "Trace service port (e.g. 14268)")

	flag.StringVar(&cfg.DBHost, "db-host", "", "Database host (e.g. localhost)")
	flag.StringVar(&cfg.DBPort, "db-port", "", "Database port (e.g. 5432)")
	flag.StringVar(&cfg.DBUser, "db-user", "", "Database user")
	flag.StringVar(&cfg.Password, "db-password", "", "Database password")
	flag.StringVar(&cfg.DBName, "db-name", "", "Database name")
	flag.StringVar(&cfg.SSLMode, "db-sslmode", "", "Database SSL mode (e.g. disable, require, verify-full)")

	flag.Parse()

	if cfg.DBHost == "" || cfg.Host == "" || cfg.TraceHost == "" {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to loading .env file : %v", err)
		}
		if err := env.Parse(&cfg); err != nil {
			return nil, fmt.Errorf("failed to parsing cfg: %v", err)
		}
		return &cfg, nil
	}
	return &cfg, nil
}

func GetDataSourceName(cfg *Config) string {
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.Password, cfg.DBName, cfg.SSLMode)

	return dataSourceName
}
