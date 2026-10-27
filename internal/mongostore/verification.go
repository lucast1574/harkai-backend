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
	filter := bson.M{"_id": id, "user_id": bson.M{"$ne": actor}, "verified_by": bson.M{"$ne": actor}, "confirmed_by": bson.M{"$ne": actor}, "status": "active", "created_at": bson.M{"$gt": now.Add(-core.AlertLifetime)},
		"$or": bson.A{bson.M{"expires_at": nil}, bson.M{"expires_at": bson.M{"$gt": now}}}}
	// One atomic update prevents duplicate votes and migrates legacy confirmations.
	update := mongo.Pipeline{bson.D{{Key: "$set", Value: bson.M{
		"confirmations": bson.M{"$add": bson.A{bson.M{"$ifNull": bson.A{"$confirmations", bson.M{"$cond": bson.A{"$verified", 1, 0}}}}, 1}},
		"confirmed_by":  bson.M{"$setUnion": bson.A{bson.M{"$ifNull": bson.A{"$confirmed_by", bson.A{}}}, bson.A{actor}}},
		"verified":      true, "verification": "community", "updated_at": now,
		"verified_by": bson.M{"$ifNull": bson.A{"$verified_by", actor}},
		"verified_at": bson.M{"$ifNull": bson.A{"$verified_at", now}},
	}}}}
	var i core.Incident
	err := s.DB.Collection("incidents").FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&i)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return core.Incident{}, core.ErrConflict
	}
	return i, translate(err)
}
