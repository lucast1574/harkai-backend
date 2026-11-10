package httpapi_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
)

func TestMunicipalInquiryStoredWithSingleAdminNotification(t *testing.T) {
	api := fixture(t)
	draft := core.MunicipalInquiry{RequestID: uuid.NewString(), Municipality: "Municipalidad de prueba", Name: "Ana Pérez", Role: "Gestión municipal", Email: "contact@example.com", Message: "Queremos coordinar una demostración del acceso municipal.", Consent: true}
	request(t, api, "POST", "/v1/municipal-inquiries", "", draft, 503)
	api.Service.MunicipalContactEnabled = true
	w := request(t, api, "POST", "/v1/municipal-inquiries", "", draft, 201)
	if strings.Contains(w.Body.String(), draft.Email) || strings.Contains(w.Body.String(), draft.Message) {
		t.Fatal("private details leaked in public receipt")
	}
	var receipt struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	request(t, api, "POST", "/v1/municipal-inquiries", "", draft, 201)
	store := api.Service.Users.(*mongostore.Store)
	count, err := store.DB.Collection("support_tickets").CountDocuments(context.Background(), bson.M{"kind": "municipal"})
	if err != nil || count != 1 {
		t.Fatal(count, err)
	}
	ticket, err := store.Ticket(context.Background(), receipt.ID, "", false)
	if err != nil || ticket.Mail["municipal_inquiry"].State != "pending" || ticket.RequesterEmail != draft.Email {
		t.Fatal(ticket, err)
	}
	request(t, api, "GET", "/v1/admin/support/tickets?kind=municipal", "", nil, 401)
	ordinary := registered(t, api, "neighbor@example.com")
	request(t, api, "GET", "/v1/admin/support/tickets?kind=municipal", ordinary.AccessToken, nil, 403)
	admin := registered(t, api, "admin@example.com")
	if _, err = store.DB.Collection("users").UpdateOne(context.Background(), bson.M{"_id": admin.User.ID}, bson.M{"$set": bson.M{"role": "admin"}}); err != nil {
		t.Fatal(err)
	}
	w = request(t, api, "GET", "/v1/admin/support/tickets?kind=municipal", admin.AccessToken, nil, 200)
	if !strings.Contains(w.Body.String(), draft.Email) || !strings.Contains(w.Body.String(), "Gestión municipal") {
		t.Fatal("admin cannot see contact")
	}
	request(t, api, "PATCH", "/v1/admin/support/tickets/"+receipt.ID, admin.AccessToken, map[string]any{"version": 0, "state": "DONE", "note": "Se coordinó una llamada."}, 200)
	ticket, err = store.Ticket(context.Background(), receipt.ID, "", false)
	if err != nil || ticket.Mail["ticket_resolved"].State != "" {
		t.Fatal("sales follow-up must not send technical-resolution email", err)
	}
}
