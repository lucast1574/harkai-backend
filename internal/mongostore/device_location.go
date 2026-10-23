package mongostore

import (
	"context"
	"time"

	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) MoveDevice(ctx context.Context, user, session, id string, point core.Point, now, expiry time.Time) (core.Device, error) {
	filter := bson.M{"_id": id, "user_id": user, "session_id": session, "expires_at": bson.M{"$gt": now}}
	update := bson.M{"$set": bson.M{"location": point, "location_updated_at": now, "expires_at": expiry}}
	var result core.Device
	err := s.DB.Collection("devices").FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&result)
	return result, translate(err)
}
