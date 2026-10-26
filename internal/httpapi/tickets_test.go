package httpapi_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestTechnicalTicketsOwnershipAtomicMailAndResolution(t *testing.T) {
	api := fixture(t)
	owner, other, admin := registered(t, api, "owner@example.com"), registered(t, api, "other@example.com"), registered(t, api, "admin@example.com")
	store := api.Service.Users.(*mongostore.Store)
	if _, err := store.DB.Collection("users").UpdateOne(context.Background(), bson.M{"_id": admin.User.ID}, bson.M{"$set": bson.M{"role": "admin"}}); err != nil {
		t.Fatal(err)
	}
	var screenshot bytes.Buffer
	if err := png.Encode(&screenshot, image.NewRGBA(image.Rect(0, 0, 3, 3))); err != nil {
		t.Fatal(err)
	}
	draft := map[string]any{"requestId": uuid.NewString(), "title": "La pantalla no carga", "description": "El perfil queda en blanco después de abrirlo.", "images": []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(screenshot.Bytes())}}
	request(t, api, "POST", "/v1/support/tickets", "", draft, 401)
	w := request(t, api, "POST", "/v1/support/tickets", owner.AccessToken, draft, 201)
	var ticket core.SupportTicket
	if err := json.Unmarshal(w.Body.Bytes(), &ticket); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(w.Body.String(), "requesterEmail") || strings.Contains(w.Body.String(), "actorId") || strings.Contains(w.Body.String(), "mail") {
		t.Fatal("private ticket fields leaked")
	}
	request(t, api, "POST", "/v1/support/tickets", owner.AccessToken, draft, 201)
	if count, err := store.DB.Collection("support_tickets").CountDocuments(context.Background(), bson.M{}); err != nil || count != 1 {
		t.Fatal("idempotent submission failed", err, count)
	}
	stored, err := store.Ticket(context.Background(), ticket.ID, "", false)
	if err != nil || stored.Mail["ticket_received"].State != "pending" {
		t.Fatal("ticket mail was not atomically queued")
	}
	request(t, api, "GET", "/v1/admin/support/tickets", other.AccessToken, nil, 403)
	request(t, api, "GET", "/v1/support/tickets/"+ticket.ID+"/images/0", other.AccessToken, nil, 404)
	w = request(t, api, "GET", "/v1/support/tickets/"+ticket.ID+"/images/0", owner.AccessToken, nil, 200)
	if !bytes.Equal(w.Body.Bytes(), screenshot.Bytes()) || !strings.Contains(strings.Join(w.Header().Values("Cache-Control"), ","), "private") {
		t.Fatal("private image response invalid")
	}
	w = request(t, api, "GET", "/v1/support/tickets", other.AccessToken, nil, 200)
	if !strings.Contains(w.Body.String(), `"total":0`) {
		t.Fatal("another account can see the ticket")
	}
	request(t, api, "PATCH", "/v1/admin/support/tickets/"+ticket.ID, owner.AccessToken, map[string]any{"version": 0, "state": "DONE", "note": "Solved"}, 403)
	request(t, api, "PATCH", "/v1/admin/support/tickets/"+ticket.ID, admin.AccessToken, map[string]any{"version": 0, "state": "DONE", "note": "Corregido"}, 200)
	stored, err = store.Ticket(context.Background(), ticket.ID, "", false)
	if err != nil || stored.Mail["ticket_resolved"].State != "pending" {
		t.Fatal("resolution mail missing")
	}
	request(t, api, "PATCH", "/v1/admin/support/tickets/"+ticket.ID, admin.AccessToken, map[string]any{"version": 1, "state": "OPEN", "note": "Reabrir"}, 400)
}

func TestConcurrentTicketUpdatesUsePersistedVersion(t *testing.T) {
	api := fixture(t)
	user := registered(t, api, "owner@example.com")
	store := api.Service.Users.(*mongostore.Store)
	if _, err := store.DB.Collection("users").UpdateOne(context.Background(), bson.M{"_id": user.User.ID}, bson.M{"$set": bson.M{"role": "admin"}}); err != nil {
		t.Fatal(err)
	}
	ticket, err := api.Service.CreateSupportTicket(context.Background(), user.User, uuid.NewString(), "Test ticket", "Enough detail to reproduce the bug", func() ([]core.SupportImage, error) { return nil, nil })
	if err != nil {
		t.Fatal(err)
	}
	user.User.Role = "admin"
	var workers sync.WaitGroup
	results := make(chan error, 2)
	for _, state := range []string{"IN_PROGRESS", "WAITING"} {
		workers.Add(1)
		go func(state string) {
			defer workers.Done()
			_, err := api.Service.UpdateSupportTicket(context.Background(), user.User, ticket.ID, state, "Investigating", 0)
			results <- err
		}(state)
	}
	workers.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatal("concurrent updates bypassed version check")
	}
}
