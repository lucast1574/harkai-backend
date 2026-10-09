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

func (s *Store) SaveDevice(ctx context.Context, d core.Device) (core.Device, error) {
	// An existing token cannot be moved to another account by an upsert.
	var result core.Device
	err := s.DB.Collection("devices").FindOneAndUpdate(ctx, bson.M{"_id": d.ID, "user_id": d.UserID}, bson.M{"$set": d}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&result)
	return result, translate(err)
}
func (s *Store) DeleteDevice(ctx context.Context, user, id string) error {
	_, err := s.DB.Collection("devices").DeleteOne(ctx, bson.M{"_id": id, "user_id": user})
	return err
}
func (s *Store) PushDevices(ctx context.Context, i core.Incident, cursor string, now time.Time) ([]core.Device, error) {
	filter := bson.M{"expires_at": bson.M{"$gt": now}, "updated_at": bson.M{"$lte": i.CreatedAt}, "location": bson.M{"$geoWithin": bson.M{"$centerSphere": bson.A{bson.A{i.Longitude, i.Latitude}, 10000.0 / 6371008.8}}}}
	if cursor != "" {
		filter["_id"] = bson.M{"$gt": cursor}
	}
	c, err := s.DB.Collection("devices").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(50))
	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)
	out := []core.Device{}
	err = c.All(ctx, &out)
	return out, err
}
func (s *Store) LiveDevice(ctx context.Context, d core.Device, now time.Time) (core.User, error) {
	count, err := s.DB.Collection("devices").CountDocuments(ctx, bson.M{"_id": d.ID, "generation": d.Generation, "session_id": d.SessionID, "expires_at": bson.M{"$gt": now}})
	if err != nil {
		return core.User{}, err
	}
	if count != 1 {
		return core.User{}, core.ErrNotFound
	}
	var session core.Session
	err = s.DB.Collection("sessions").FindOne(ctx, bson.M{"_id": d.SessionID, "user_id": d.UserID, "expires_at": bson.M{"$gt": now}}).Decode(&session)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return core.User{}, core.ErrNotFound
	}
	if err != nil {
		return core.User{}, err
	}
	return s.UserByID(ctx, d.UserID)
}
func (s *Store) DisableDevice(ctx context.Context, d core.Device) error {
	_, err := s.DB.Collection("devices").DeleteOne(ctx, bson.M{"_id": d.ID, "generation": d.Generation})
	return err
}
