package runtime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/lucast1574/harkai-backend/internal/core"
	email "github.com/lucast1574/harkai-backend/internal/mail"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type municipalTransport func(*http.Request) (*http.Response, error)

func (f municipalTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMunicipalNotificationRetriesToConfiguredAdminOnly(t *testing.T) {
	store := testutil.Store(t)
	now := UTCNow()
	ticket := core.SupportTicket{ID: mongostore.NewID(), Kind: "municipal", RequesterID: "municipal", RequestID: "email-test", RequesterName: "Ana", RequesterEmail: "contact@example.com", Title: "Acceso municipal · Municipalidad de prueba", Description: "Queremos coordinar una llamada sobre nuestro distrito.", State: "OPEN", CreatedAt: now, UpdatedAt: now, Mail: map[string]core.EmailDelivery{"municipal_inquiry": core.PendingEmail(now)}}
	if _, err := store.CreateTicket(context.Background(), ticket); err != nil {
		t.Fatal(err)
	}
	original := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = original })
	calls := 0
	http.DefaultClient = &http.Client{Transport: municipalTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		var body struct {
			To      []string `json:"to"`
			Subject string   `json:"subject"`
			HTML    string   `json:"html"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.To) != 1 || body.To[0] != "owner@example.com" {
			t.Fatal("notification sent to wrong recipient", body.To)
		}
		if !strings.Contains(body.Subject, "Municipalidad de prueba") || !strings.Contains(body.HTML, ticket.Description) || !strings.Contains(body.HTML, "section=municipal") {
			t.Fatal("notification lost inquiry context")
		}
		if r.Header.Get("Idempotency-Key") != "harkai:"+ticket.ID+":municipal_inquiry" {
			t.Fatal("unstable delivery key")
		}
		status := 500
		if calls > 1 {
			status = 200
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(`{"id":"notification"}`)), Header: http.Header{}}, nil
	})}
	worker := emailWorker{store: store, config: email.Config{APIKey: "fixture", From: "Harkai <hello@example.com>", MunicipalTo: "owner@example.com"}}
	if err := worker.deliver(context.Background(), "support_tickets", "municipal_inquiry"); err != nil {
		t.Fatal(err)
	}
	saved, err := store.Ticket(context.Background(), ticket.ID, "", false)
	if err != nil || saved.Mail["municipal_inquiry"].State != "pending" {
		t.Fatal("failed notification not retried", err)
	}
	if _, err = store.DB.Collection("support_tickets").UpdateOne(context.Background(), bson.M{"_id": ticket.ID}, bson.M{"$set": bson.M{"mail.municipal_inquiry.nextAt": now.Add(-time.Minute)}}); err != nil {
		t.Fatal(err)
	}
	worker.config.MunicipalTo = "changed@example.com"
	if err = worker.deliver(context.Background(), "support_tickets", "municipal_inquiry"); err != nil {
		t.Fatal(err)
	}
	saved, err = store.Ticket(context.Background(), ticket.ID, "", false)
	if err != nil || saved.Mail["municipal_inquiry"].State != "sent" || calls != 2 {
		t.Fatal("notification retry failed", err)
	}
}
