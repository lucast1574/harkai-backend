package mongostore

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func (s *Store) ConfirmIncident(ctx context.Context, id, actor string, now time.Time) (core.Incident, error) {
	if actor == "" {
		return core.Incident{}, core.ErrUnauthorized
	}
	filter := bson.M{"_id": id, "user_id": bson.M{"$ne": actor}, "verified": bson.M{"$ne": true}, "status": "active",
		"$or": bson.A{bson.M{"expires_at": nil}, bson.M{"expires_at": bson.M{"$gt": now}}}}
	update := bson.M{"$set": bson.M{"verified": true, "verified_by": actor, "verified_at": now, "verification": "community", "updated_at": now}}
	var i core.Incident
	err := s.DB.Collection("incidents").FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&i)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return core.Incident{}, core.ErrConflict
	}
	return i, translate(err)
}
