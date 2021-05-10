package main

import (
	"context"
	"fsn/config"
	"fsn/controller"
	"fsn/logger"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = setLog(ctx)

	cfg := config.LoadConfig(ctx)

	cfg.PID = os.Getpid()
	cfg.Save(ctx)

	ctrl := controller.NewController(cfg)
	ctrl.Run(ctx, cfg.Timeout, cfg)

	sigs := make(chan os.Signal, 1000)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGHUP)

	logger.Info(ctx, "Service is running.")

	for s := range sigs {
		switch s {
		case syscall.SIGTERM:
			logger.Info(ctx, "Getting exit signal, bye!")
			cancel()
			os.Exit(1)
		case syscall.SIGHUP:
			logger.Info(ctx, "Getting reload signal, checking for changes...")
			newCfg := config.LoadConfig(ctx)
			ctrl.Reload(ctx, newCfg)
			logger.Info(ctx, "Reload successful")
		}
	}
}

func setLog(ctx context.Context) context.Context {
	fpath := logger.GetFilepath()

	err := os.MkdirAll(path.Dir(fpath), 0777)
	if err != nil {
		log.Println(err)
	}

	logWriter, err := os.Create(fpath)
	if err != nil {
		log.Println(err)
	}

	return logger.NewLogger(ctx, logWriter)
}
