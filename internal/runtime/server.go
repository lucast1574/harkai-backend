package runtime

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/config"
	"github.com/lucast1574/harkai-backend/internal/districtgeo"
	"github.com/lucast1574/harkai-backend/internal/googleauth"
	"github.com/lucast1574/harkai-backend/internal/httpapi"
	"github.com/lucast1574/harkai-backend/internal/mediatools"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/push"
	"github.com/lucast1574/harkai-backend/internal/s3store"
	"github.com/lucast1574/harkai-backend/internal/security"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func Run(ctx context.Context, c config.Config) error {
	ctx, stop := context.WithCancel(ctx)
	defer stop()
	if err := districtgeo.Warm(); err != nil {
		return errors.New("district reference catalog failed")
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	startup, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	store, err := mongostore.Connect(startup, c.MongoURI, c.Database)
	if err != nil {
		return err
	}
	defer closeStore(store)
	if err = store.EnsureIndexes(startup); err != nil {
		return errors.New("database indexes failed")
	}
	if err = store.SeedSupport(startup); err != nil {
		return errors.New("support directory seed failed")
	}
	if err = store.EnsureAdmins(startup); err != nil {
		return errors.New("admin role sync failed")
	}
	photos, err := s3store.Connect(startup, c.S3Endpoint, c.S3AccessKey, c.S3SecretKey, c.S3Bucket, c.S3Secure)
	if err != nil {
		return err
	}
	service, err := NewService(store, photos)
	if err != nil {
		return err
	}
	if c.MediaWorkerURL != "" {
		worker, err := mediatools.New(c.MediaWorkerURL, c.MediaWorkerKey)
		if err != nil {
			return err
		}
		service.ModerateImage, service.TranscribeAudio, service.MediaSource = worker.Moderate, worker.Transcribe, worker.Source
	}
	if c.GoogleProject != "" {
		service.VerifyGoogle = googleauth.New(c.GoogleProject).Verify
	}
	ready := func(ctx context.Context) error {
		if err := store.Ping(ctx); err != nil {
			return err
		}
		return photos.Ping(ctx)
	}
	if c.FCMKeyBase64 != "" {
		raw, err := base64.StdEncoding.DecodeString(c.FCMKeyBase64)
		if err != nil {
			return errors.New("invalid FCM credential encoding")
		}
		sender, err := push.New(ctx, c.GoogleProject, raw)
		if err != nil {
			return err
		}
		validation, cancelValidation := context.WithTimeout(ctx, 15*time.Second)
		err = sender.Validate(validation)
		cancelValidation()
		if err != nil {
			return err
		}
		service.PushEnabled = true
		worker := push.Worker{Store: store, Sender: sender, Now: UTCNow, ID: mongostore.NewID}
		stopWorker := startPush(ctx, worker, logger)
		defer stopWorker()
	}
	handler := httpapi.New(service, ready, c.Origins, logger)
	server := &http.Server{Addr: ":" + c.Port, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 50 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	logger.Info("Harkai API listening", "port", c.Port)
	return Serve(ctx, server)
}

func NewService(store *mongostore.Store, photos application.PhotoStore) (application.Service, error) {
	dummy, err := security.HashPassword("unused-dummy-password-for-timing")
	if err != nil {
		return application.Service{}, err
	}
	return application.Service{Forum: store, DistrictByID: districtgeo.ByID, Support: store, Districts: store, Devices: store, Comments: store, Users: store, Sessions: store, Incidents: store, Limits: store, Media: store, Photos: photos, Directory: store, NormalizeImage: security.NormalizeImage, Now: UTCNow, ID: mongostore.NewID, Token: security.Token, HashPassword: security.HashPassword, CheckPassword: security.CheckPassword, HashToken: security.HashToken, DummyPasswordHash: dummy}, nil
}

func UTCNow() time.Time { return time.Now().UTC() }

func closeStore(store *mongostore.Store) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return store.Close(ctx)
}

func Serve(ctx context.Context, server *http.Server) error {
	done := make(chan error, 1)
	go listen(server, done)
	select {
	case err := <-done:
		return normalizeServerError(err)
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := server.Shutdown(shutdown)
		if err != nil {
			_ = server.Close()
			return err
		}
		return normalizeServerError(<-done)
	}
}

func listen(server *http.Server, done chan<- error) error {
	err := server.ListenAndServe()
	done <- err
	return err
}

func normalizeServerError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
