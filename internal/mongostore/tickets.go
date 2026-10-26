package mongostore

import (
	"context"
	"errors"

	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) PriorTicket(ctx context.Context, user, request string) (core.SupportTicket, error) {
	ticket, err := (SupportTickets{DB: s.DB}).Prior(ctx, user, request)
	return ticket, translate(err)
}
func (s *Store) CreateTicket(ctx context.Context, ticket core.SupportTicket) (core.SupportTicket, error) {
	result, err := (SupportTickets{DB: s.DB}).Create(ctx, ticket)
	return result, translate(err)
}
func (s *Store) ListTickets(ctx context.Context, query application.TicketQuery) ([]core.SupportTicket, int64, error) {
	items, total, err := (SupportTickets{DB: s.DB}).List(ctx, SupportFilter{RequesterID: query.RequesterID, State: query.State}, query.Page)
	return items, total, translate(err)
}
func (s *Store) Ticket(ctx context.Context, id, owner string, images bool) (core.SupportTicket, error) {
	ticket, err := (SupportTickets{DB: s.DB}).Get(ctx, id, owner, images)
	return ticket, translate(err)
}
func (s *Store) ChangeTicket(ctx context.Context, previous core.SupportTicket, change core.SupportChange) (core.SupportTicket, error) {
	ticket, err := (SupportTickets{DB: s.DB}).Change(ctx, previous, change)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ticket, core.ErrConflict
	}
	return ticket, translate(err)
}
func (s *Store) EnsureTicketIndexes(ctx context.Context) error {
	_, err := s.DB.Collection("support_tickets").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "requesterId", Value: 1}, {Key: "requestId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "requesterId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "state", Value: 1}, {Key: "createdAt", Value: -1}}},
	})
	if err != nil {
		return err
	}
	return EnsureEmailIndexes(ctx, s.DB)
}
