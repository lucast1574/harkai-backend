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
		page.Items[n] = s.PublicationIncident(page.Items[n])
	}
	return page, err
}

// Resolve missing legacy location labels from the bundled district boundaries.
func (s Service) PublicationIncident(i core.Incident) core.Incident {
	if s.DistrictAt != nil && (i.District == "" || i.City == "") {
		if district, err := s.DistrictAt(i.Latitude, i.Longitude); err == nil {
			if i.District == "" {
				i.District = district.Name
			}
			if i.City == "" {
				i.City = district.Province
			}
			if i.Country == "" {
				i.Country = "Perú"
			}
		}
	}
	return core.PublicIncident(i)
}
