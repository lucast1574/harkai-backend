package runtime

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/lucast1574/harkai-backend/internal/core"
	email "github.com/lucast1574/harkai-backend/internal/mail"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type emailWorker struct {
	store  *mongostore.Store
	config email.Config
	logger *slog.Logger
}

func startEmails(ctx context.Context, store *mongostore.Store, config email.Config, logger *slog.Logger) func() error {
	if !config.Valid() {
		return func() error { return nil }
	}
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	worker := emailWorker{store: store, config: config, logger: logger}
	go worker.run(ctx, done)
	return func() error { cancel(); return <-done }
}

func (worker emailWorker) run(ctx context.Context, done chan<- error) error {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			done <- nil
			return nil
		case <-ticker.C:
			for _, job := range [][2]string{{"users", "welcome"}, {"support_tickets", "ticket_received"}, {"support_tickets", "ticket_resolved"}} {
				if err := worker.deliver(ctx, job[0], job[1]); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
					worker.logger.Warn("Transactional email pending; delivery will retry")
				}
			}
		}
	}
}

func (worker emailWorker) deliver(ctx context.Context, collection, kind string) error {
	repo := mongostore.Emails{DB: worker.store.DB}
	lease, now := mongostore.NewID(), UTCNow()
	job, err := repo.Claim(ctx, collection, kind, lease, now)
	if err != nil {
		return err
	}
	delivery := job.Mail[kind]
	payload := delivery.Payload
	if payload == nil {
		name, address, action := job.Name, job.Email, "https://panel.harkai.lat/dashboard"
		if collection == "support_tickets" {
			name, address, action = job.RequesterName, job.RequesterEmail, "https://panel.harkai.lat/dashboard/tickets"
		}
		content, err := email.Render("general", kind, email.TemplateData{Name: name, Title: job.Title, ActionURL: action})
		if err != nil {
			return err
		}
		subject := map[string]string{"welcome": "Tu cuenta está lista", "ticket_received": "Recibimos tu ticket", "ticket_resolved": "Tu ticket está resuelto"}[kind]
		payload = &core.EmailPayload{From: worker.config.From, To: address, Subject: "Harkai · " + subject, HTML: content.HTML, Text: content.Text}
		if err = repo.Prepare(ctx, collection, kind, job.ID, lease, *payload); err != nil {
			return err
		}
	}
	config := worker.config
	config.From = payload.From
	result := email.SendResend(ctx, config, payload.To, payload.Subject, payload.HTML, payload.Text, "harkai:"+job.ID+":"+kind)
	state, delay := "sent", time.Duration(0)
	if result.State != "sent" {
		state = "pending"
		delay = time.Minute * time.Duration(1<<min(delivery.Attempts, 6))
	}
	return repo.Finish(ctx, collection, kind, job.ID, lease, state, UTCNow().Add(delay))
}
