package mongostore

import (
	"context"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func (s *Store) Allow(ctx context.Context, key string, max int, window time.Duration, now time.Time) error {
	slot := now.UnixNano() / int64(window)
	id := fmt.Sprintf("%s:%d", key, slot)
	end := time.Unix(0, (slot+1)*int64(window)).Add(time.Hour)
	_, err := s.DB.Collection("rate_limits").UpdateOne(ctx, bson.M{"_id": id, "count": bson.M{"$lt": max}}, bson.M{"$inc": bson.M{"count": 1}, "$setOnInsert": bson.M{"expires_at": end}}, options.UpdateOne().SetUpsert(true))
	if mongo.IsDuplicateKeyError(err) {
		return core.ErrRateLimited
	}
	return err
}

func (s *Store) ClaimCooldown(ctx context.Context, userID, kind string, now time.Time) error {
	_, err := s.DB.Collection("cooldowns").UpdateOne(ctx, bson.M{"_id": userID + ":" + kind, "next_at": bson.M{"$lte": now}}, bson.M{"$set": bson.M{"next_at": now.Add(time.Hour)}}, options.UpdateOne().SetUpsert(true))
	if mongo.IsDuplicateKeyError(err) {
		return core.ErrRateLimited
	}
	return err
}
