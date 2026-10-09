package runtime

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/lucast1574/harkai-backend/internal/push"
)

func startPush(ctx context.Context, worker push.Worker, logger *slog.Logger) func() error {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go runPush(ctx, worker, logger, done)
	return func() error {
		cancel()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		select {
		case <-done:
			return nil
		case <-timer.C:
			return errors.New("notification worker shutdown timed out")
		}
	}
}

func runPush(ctx context.Context, worker push.Worker, logger *slog.Logger, done chan<- error) error {
	err := worker.Run(ctx, logger)
	done <- err
	return err
}
