package main

import (
	"os"
	"os/signal"
	"syscall"
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

	// storage init

	// application init

	// start servers

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)
	<-done

	// close connections
}
