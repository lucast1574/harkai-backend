package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) EmergencyContacts(ctx context.Context, country, city string) (core.EmergencyContacts, error) {
	var contacts core.EmergencyContacts
	err := s.DB.Collection("emergency_contacts").FindOne(ctx, bson.M{"_id": country + ":" + city}).Decode(&contacts)
	return contacts, translate(err)
}

func (s *Store) SaveEmergencyContacts(ctx context.Context, c core.EmergencyContacts) (core.EmergencyContacts, error) {
	_, err := s.DB.Collection("emergency_contacts").ReplaceOne(ctx, bson.M{"_id": c.ID}, c, options.Replace().SetUpsert(true))
	return c, translate(err)
}
