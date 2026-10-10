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
	_, city, err := core.Region("PE", city)
	if err != nil {
		return nil, err
	}
	if city != "lima" && city != "trujillo" {
		return nil, fmt.Errorf("%w: unsupported support city", core.ErrInvalid)
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
