package main

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/config"
	"github.com/lucast1574/harkai-backend/internal/runtime"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		slog.Error("startup failed", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	c, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runtime.Run(ctx, c)
}
