package mongostore

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) EnsureIndexes(ctx context.Context) error {
	indexes := map[string][]mongo.IndexModel{
		"users": {{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)}, {Keys: bson.D{{Key: "google_subject", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)}},
		"sessions": {
			{Keys: bson.D{{Key: "access_hash", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "refresh_hash", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
		},
		"incidents": {
			{Keys: bson.D{{Key: "location", Value: "2dsphere"}}},
			{Keys: bson.D{{Key: "status", Value: 1}, {Key: "type", Value: 1}, {Key: "_id", Value: -1}}},
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "_id", Value: -1}}},
			{Keys: bson.D{{Key: "media_id", Value: 1}}},
		},
		"rate_limits": {{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)}},
		"cooldowns":   {{Keys: bson.D{{Key: "next_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)}},
		"media":       {{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}}},
	}
	for collection, models := range indexes {
		if _, err := s.DB.Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
			return err
		}
	}
	return nil
}
