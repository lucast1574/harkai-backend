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
	filter := bson.M{"_id": n.ID, "$or": bson.A{bson.M{"next_eligible_at": bson.M{"$lte": n.CreatedAt}}, bson.M{"stage": bson.M{"$gt": n.Stage}}}}
	_, err := s.DB.Collection("business_notices").ReplaceOne(ctx, filter, n, options.Replace().SetUpsert(true))
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}
func (s *Store) BusinessNotices(ctx context.Context, user, session, device string, now time.Time) ([]core.BusinessNotice, error) {
	rows, err := s.DB.Collection("business_notices").Find(ctx, bson.M{"user_id": user, "session_id": session, "device_id": device, "pending": true, "created_at": bson.M{"$gt": now.Add(-24 * time.Hour)}}, options.Find().SetSort(bson.D{{Key: "kind", Value: -1}, {Key: "created_at", Value: 1}}).SetLimit(50))
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
		valid := false
		if n.Kind == "incident" {
			i, e := s.IncidentByID(ctx, n.PointID)
			if e != nil && e != core.ErrNotFound {
				return nil, e
			}
			valid = e == nil && core.MapIncidentActive(i, now)
		} else {
			p, e := s.BusinessByID(ctx, n.PointID)
			if e != nil && e != core.ErrNotFound {
				return nil, e
			}
			valid = e == nil && p.Public(now)
		}
		if valid {
			result = append(result, n)
		} else {
			if err = s.AckProximityNotice(ctx, user, session, n.ID, n.Stage); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}
func (s *Store) AckBusinessNotice(ctx context.Context, user, session, id string) error {
	return s.AckProximityNotice(ctx, user, session, id, 0)
}
func (s *Store) AckProximityNotice(ctx context.Context, user, session, id string, stage int) error {
	filter := bson.M{"_id": id, "user_id": user, "session_id": session, "stage": stage}
	if stage == 0 {
		delete(filter, "stage")
		filter["$or"] = bson.A{bson.M{"stage": 0}, bson.M{"stage": bson.M{"$exists": false}}}
	}
	_, err := s.DB.Collection("business_notices").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"pending": false}})
	return err
}
