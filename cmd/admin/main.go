package main

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/config"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/security"
	"log/slog"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("admin provisioning failed", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("admin account created")
}

func run() error {
	c, err := config.Load()
	if err != nil {
		return err
	}
	input, err := core.ValidateRegistration(core.Credentials{Email: os.Getenv("ADMIN_EMAIL"), Name: os.Getenv("ADMIN_NAME"), Password: os.Getenv("ADMIN_PASSWORD")})
	if err != nil {
		return err
	}
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	store, err := mongostore.Connect(ctx, c.MongoURI, c.Database)
	if err != nil {
		return err
	}
	defer store.Close(ctx)
	if err = store.EnsureIndexes(ctx); err != nil {
		return errors.New("database indexes failed")
	}
	_, err = store.CreateUser(ctx, core.User{ID: mongostore.NewID(), Email: input.Email, Name: input.Name, Role: "admin", PasswordHash: hash, Preferences: core.DefaultPreferences(), CreatedAt: time.Now().UTC()})
	if err != nil {
		return errors.New("admin account could not be created; existing accounts are not modified")
	}
	return nil
}
