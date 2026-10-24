package runtime

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"log/slog"
	"time"
)

func startAccountDeletion(ctx context.Context, service application.Service, photos application.PhotoEraser, logger *slog.Logger) func() error {
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() error {
		defer close(done)
		timer := time.NewTicker(15 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return nil
			case <-timer.C:
				step, stop := context.WithTimeout(workerCtx, 45*time.Second)
				err := service.ProcessAccountDeletion(step, photos)
				stop()
				if err != nil && !errors.Is(err, core.ErrNotFound) && !errors.Is(err, context.Canceled) {
					logger.Error("account deletion will retry", "error", err.Error())
				}
			}
		}
	}()
	return func() error {
		cancel()
		select {
		case <-done:
			return nil
		case <-time.After(5 * time.Second):
			return errors.New("account deletion worker shutdown timed out")
		}
	}
}
