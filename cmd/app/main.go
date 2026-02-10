package main

import (
	"os"
	"os/signal"
	"syscall"
	"usersservice/internal/app"
	"usersservice/internal/storage/postgres"
	"usersservice/pkg/config"
	"usersservice/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log, err := logger.Setup(cfg.Env)
	if err != nil {
		panic("failed to setup logger: " + err.Error())
	}

	log.Info("application trying to setting up", zap.Any("env", cfg.Env))

	storage, close, err := postgres.New(log, cfg.Postgres.DSN())
	if err != nil {
		log.Fatal("failed to setup storage", zap.Error(err))
	}
	_ = storage

	application := app.New(log, storage, cfg.Rest.Port)
	go func() {
		if err := application.RESTServer.Start(); err != nil {
			log.Fatal("failed to start REST server", zap.Error(err))
		}
	}()
	// TODO: add gRPC server

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)
	<-done

	close()
	application.RESTServer.Shutdown()
}
