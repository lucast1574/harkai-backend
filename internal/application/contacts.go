package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
)

type ContactsStore interface {
	EmergencyContacts(context.Context, string, string) (core.EmergencyContacts, error)
	SaveEmergencyContacts(context.Context, core.EmergencyContacts) (core.EmergencyContacts, error)
}

func (s Service) Contacts(ctx context.Context, country, city string) (core.EmergencyContacts, error) {
	country, city, err := core.Region(country, city)
	if err != nil {
		return core.EmergencyContacts{}, err
	}
	return s.Directory.EmergencyContacts(ctx, country, city)
}

func (s Service) SaveContacts(ctx context.Context, user core.User, input core.EmergencyContacts) (core.EmergencyContacts, error) {
	if user.Role != "admin" {
		return core.EmergencyContacts{}, core.ErrForbidden
	}
	input, err := core.ValidateContacts(input)
	if err != nil {
		return core.EmergencyContacts{}, err
	}
	return s.Directory.SaveEmergencyContacts(ctx, input)
}
