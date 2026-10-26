package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lucast1574/harkai-backend/internal/core"
)

type TicketQuery struct {
	RequesterID, State string
	Page               int64
}
type TicketStore interface {
	PriorTicket(context.Context, string, string) (core.SupportTicket, error)
	CreateTicket(context.Context, core.SupportTicket) (core.SupportTicket, error)
	ListTickets(context.Context, TicketQuery) ([]core.SupportTicket, int64, error)
	Ticket(context.Context, string, string, bool) (core.SupportTicket, error)
	ChangeTicket(context.Context, core.SupportTicket, core.SupportChange) (core.SupportTicket, error)
}

func (s Service) CreateSupportTicket(ctx context.Context, user core.User, requestID, title, description string, decodeImages func() ([]core.SupportImage, error)) (core.SupportTicket, error) {
	request, err := uuid.Parse(requestID)
	if err != nil {
		return core.SupportTicket{}, core.ErrInvalid
	}
	title, description, err = core.SupportText(title, description)
	if err != nil {
		return core.SupportTicket{}, fmt.Errorf("%w: ticket text", core.ErrInvalid)
	}
	previous, err := s.Tickets.PriorTicket(ctx, user.ID, request.String())
	if err == nil {
		return previous, nil
	}
	if !errors.Is(err, core.ErrNotFound) {
		return core.SupportTicket{}, err
	}
	now := s.Now()
	if err = s.Limits.Allow(ctx, "tickets:"+user.ID, 5, time.Minute, now); err != nil {
		return core.SupportTicket{}, err
	}
	images, err := decodeImages()
	if err != nil || len(images) > 3 {
		return core.SupportTicket{}, core.ErrInvalid
	}
	ticket := core.SupportTicket{ID: s.ID(), RequestID: request.String(), RequesterID: user.ID, RequesterName: user.Name, RequesterEmail: user.Email, Title: title, Description: description, State: "OPEN", Images: images, History: []core.SupportChange{{State: "OPEN", At: now, ActorID: user.ID}}, CreatedAt: now, UpdatedAt: now, Mail: map[string]core.EmailDelivery{"ticket_received": core.PendingEmail(now)}}
	return s.Tickets.CreateTicket(ctx, ticket)
}

func (s Service) SupportTickets(ctx context.Context, user core.User, admin bool, query TicketQuery) ([]core.SupportTicket, int64, error) {
	if admin && user.Role != "admin" {
		return nil, 0, core.ErrForbidden
	}
	if query.State != "" && !core.ValidSupportState(query.State) {
		return nil, 0, core.ErrInvalid
	}
	if query.Page < 1 || query.Page > 100000 {
		return nil, 0, core.ErrInvalid
	}
	query.RequesterID = ""
	if !admin {
		query.RequesterID = user.ID
	}
	return s.Tickets.ListTickets(ctx, query)
}

func (s Service) SupportImage(ctx context.Context, user core.User, id string, index int) (core.SupportImage, error) {
	if index < 0 || index > 2 {
		return core.SupportImage{}, core.ErrNotFound
	}
	owner := user.ID
	if user.Role == "admin" {
		owner = ""
	}
	ticket, err := s.Tickets.Ticket(ctx, id, owner, true)
	if err != nil {
		return core.SupportImage{}, err
	}
	if index >= len(ticket.Images) {
		return core.SupportImage{}, core.ErrNotFound
	}
	return ticket.Images[index], nil
}

func (s Service) UpdateSupportTicket(ctx context.Context, user core.User, id, state, note string, version int64) (core.SupportTicket, error) {
	if user.Role != "admin" {
		return core.SupportTicket{}, core.ErrForbidden
	}
	if version < 0 {
		return core.SupportTicket{}, core.ErrInvalid
	}
	previous, err := s.Tickets.Ticket(ctx, id, "", false)
	if err != nil {
		return core.SupportTicket{}, err
	}
	if previous.Version != version {
		return core.SupportTicket{}, core.ErrConflict
	}
	change, err := core.SupportTransition(previous, state, strings.TrimSpace(note), user.ID, s.Now())
	if err != nil {
		return core.SupportTicket{}, fmt.Errorf("%w: resolved tickets are final; choose a state or add a reply", core.ErrInvalid)
	}
	return s.Tickets.ChangeTicket(ctx, previous, change)
}
