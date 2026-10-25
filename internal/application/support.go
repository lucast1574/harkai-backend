package application

import (
	"context"
	"fmt"

	"github.com/lucast1574/harkai-backend/internal/core"
)

type SupportStore interface {
	SupportDirectory(context.Context, string) ([]core.SupportDistrict, error)
	SupportPlaces(context.Context, core.IncidentQuery) ([]core.SupportPlace, error)
}

func (s Service) SupportDirectory(ctx context.Context, city string) ([]core.SupportDistrict, error) {
	if city != "lima" && city != "trujillo" {
		if len(city) != 4 || s.DistrictByID == nil {
			return nil, fmt.Errorf("%w: unknown support province", core.ErrInvalid)
		}
		if _, err := s.DistrictByID(city + "01"); err != nil {
			return nil, fmt.Errorf("%w: unknown support province", core.ErrInvalid)
		}
	}
	if s.Support == nil {
		return nil, core.ErrUnavailable
	}
	return s.Support.SupportDirectory(ctx, city)
}
func (s Service) SupportPlaces(ctx context.Context, query core.IncidentQuery) ([]core.SupportPlace, error) {
	if s.Support == nil {
		return nil, core.ErrUnavailable
	}
	return s.Support.SupportPlaces(ctx, query)
}
