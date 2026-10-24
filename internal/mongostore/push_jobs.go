package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func (s *Store) ClaimPush(ctx context.Context, lease string, now time.Time) (core.Incident, error) {
	var i core.Incident
	filter := bson.M{"push.state": "pending", "push.next_at": bson.M{"$lte": now}, "status": "active", "created_at": bson.M{"$gt": now.Add(-3 * time.Hour)}, "$or": bson.A{bson.M{"expires_at": nil}, bson.M{"expires_at": bson.M{"$gt": now}}}}
	update := bson.M{"$set": bson.M{"push.lease": lease, "push.next_at": now.Add(90 * time.Second)}, "$inc": bson.M{"push.attempts": 1}}
	err := s.DB.Collection("incidents").FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After).SetSort(bson.D{{Key: "_id", Value: 1}})).Decode(&i)
	return i, translate(err)
}
func (s *Store) AdvancePush(ctx context.Context, id, lease, cursor string, done bool, next time.Time) error {
	state := "pending"
	if done {
		state = "done"
	}
	r, err := s.DB.Collection("incidents").UpdateOne(ctx, bson.M{"_id": id, "push.lease": lease}, bson.M{"$set": bson.M{"push.cursor": cursor, "push.state": state, "push.next_at": next}})
	if err != nil {
		return err
	}
	if r.MatchedCount != 1 {
		return core.ErrConflict
	}
	return nil
}
