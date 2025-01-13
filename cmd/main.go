package main

import (
	"USDT-rate-service/cmd/config"
	"USDT-rate-service/internal/app"
	"USDT-rate-service/pkg/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger.BuildLogger(cfg.LogLevel)
	log := logger.Logger().Named("USDT_Rates_Service")

	application, err := app.NewApp(cfg, log)
	if err != nil {
		panic(err)
	}

	application.Run()
}
