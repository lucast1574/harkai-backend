package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

func (s Service) CreateMunicipalInquiry(ctx context.Context, in core.MunicipalInquiry) (core.SupportTicket, error) {
	if !s.MunicipalContactEnabled || s.Tickets == nil {
		return core.SupportTicket{}, core.ErrUnavailable
	}
	in, err := core.ValidateMunicipalInquiry(in)
	if err != nil {
		return core.SupportTicket{}, err
	}
	request, err := uuid.Parse(in.RequestID)
	if err != nil || request == uuid.Nil {
		return core.SupportTicket{}, core.ErrInvalid
	}
	previous, err := s.Tickets.PriorTicket(ctx, "municipal", request.String())
	if err == nil {
		return previous, nil
	}
	if !errors.Is(err, core.ErrNotFound) {
		return core.SupportTicket{}, err
	}
	now := s.Now()
	if err = s.Limits.Allow(ctx, "municipal-email:"+s.HashToken(in.Email), 3, 24*time.Hour, now); err != nil {
		return core.SupportTicket{}, err
	}
	if err = s.Limits.Allow(ctx, "municipal-total", 100, 24*time.Hour, now); err != nil {
		return core.SupportTicket{}, err
	}
	ticket := core.SupportTicket{
		ID: s.ID(), Kind: "municipal", RequestID: request.String(), RequesterID: "municipal",
		RequesterName: in.Name, RequesterEmail: in.Email, Title: "Acceso municipal · " + in.Municipality,
		Description: fmt.Sprintf("Municipalidad: %s\nContacto: %s\nCargo: %s\nCorreo: %s\nTeléfono: %s\n\n%s", in.Municipality, in.Name, in.Role, in.Email, in.Phone, in.Message),
		State:       "OPEN", CreatedAt: now, UpdatedAt: now, Images: []core.SupportImage{},
		History: []core.SupportChange{{State: "OPEN", At: now, ActorID: "municipal-form"}},
		Mail:    map[string]core.EmailDelivery{"municipal_inquiry": core.PendingEmail(now)},
	}
	return s.Tickets.CreateTicket(ctx, ticket)
}
