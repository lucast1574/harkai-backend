package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func (s *Store) QueueBusinessNotice(ctx context.Context, n core.BusinessNotice) error {
	filter := bson.M{"_id": n.ID, "next_eligible_at": bson.M{"$lte": n.CreatedAt}}
	_, err := s.DB.Collection("business_notices").ReplaceOne(ctx, filter, n, options.Replace().SetUpsert(true))
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}
func (s *Store) BusinessNotices(ctx context.Context, user, session, device string, now time.Time) ([]core.BusinessNotice, error) {
	rows, err := s.DB.Collection("business_notices").Find(ctx, bson.M{"user_id": user, "session_id": session, "device_id": device, "pending": true, "created_at": bson.M{"$gt": now.Add(-24 * time.Hour)}}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetLimit(3))
	if err != nil {
		return nil, err
	}
	defer rows.Close(ctx)
	items := []core.BusinessNotice{}
	if err = rows.All(ctx, &items); err != nil {
		return nil, err
	}
	result := []core.BusinessNotice{}
	for _, n := range items {
		p, e := s.BusinessByID(ctx, n.PointID)
		if e == nil && p.Public(now) {
			n.Name = p.Name
			result = append(result, n)
		} else if e == nil || e == core.ErrNotFound {
			if err = s.AckBusinessNotice(ctx, user, session, n.ID); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}
func (s *Store) AckBusinessNotice(ctx context.Context, user, session, id string) error {
	_, err := s.DB.Collection("business_notices").UpdateOne(ctx, bson.M{"_id": id, "user_id": user, "session_id": session}, bson.M{"$set": bson.M{"pending": false}})
	return err
}
