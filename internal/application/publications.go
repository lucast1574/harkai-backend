package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
)

type PublicationStore interface {
	ListPublications(context.Context, string, int) (core.IncidentPage, error)
}

func (s Service) PublicPublications(ctx context.Context, cursor string, limit int) (core.IncidentPage, error) {
	if s.Publications == nil {
		return core.IncidentPage{}, core.ErrUnavailable
	}
	if limit < 1 || limit > 100 || (cursor != "" && !core.ValidID(cursor)) {
		return core.IncidentPage{}, core.ErrInvalid
	}
	page, err := s.Publications.ListPublications(ctx, cursor, limit)
	for n := range page.Items {
		page.Items[n] = core.PublicIncident(page.Items[n])
	}
	return page, err
}
