package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func (s *Store) BusinessCheckoutByID(ctx context.Context, id string) (core.BusinessCheckout, error) {
	var record core.BusinessCheckout
	err := s.DB.Collection("business_checkouts").FindOne(ctx, bson.M{"_id": id}).Decode(&record)
	return record, translate(err)
}

func (s *Store) BusinessCheckout(ctx context.Context, candidate core.BusinessCheckout) (core.BusinessCheckout, error) {
	c := s.DB.Collection("business_checkouts")
	_, err := c.UpdateOne(ctx, bson.M{"_id": candidate.PointID}, bson.M{"$setOnInsert": candidate}, options.UpdateOne().SetUpsert(true))
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return candidate, translate(err)
	}
	var record core.BusinessCheckout
	err = c.FindOne(ctx, bson.M{"_id": candidate.PointID, "owner_id": candidate.OwnerID}).Decode(&record)
	return record, translate(err)
}
func (s *Store) RotateBusinessCheckout(ctx context.Context, old core.BusinessCheckout, key string) error {
	r, err := s.DB.Collection("business_checkouts").UpdateOne(ctx, bson.M{"_id": old.PointID, "owner_id": old.OwnerID, "key": old.Key}, bson.M{"$set": bson.M{"key": key, "session_id": "", "created_at": time.Now().UTC()}})
	if err != nil {
		return translate(err)
	}
	if r.MatchedCount == 0 {
		return core.ErrConflict
	}
	return nil
}

func (s *Store) SaveBusinessCheckout(ctx context.Context, record core.BusinessCheckout, id string) error {
	r, err := s.DB.Collection("business_checkouts").UpdateOne(ctx, bson.M{"_id": record.PointID, "key": record.Key}, bson.M{"$set": bson.M{"session_id": id}})
	if err != nil {
		return translate(err)
	}
	if r.MatchedCount == 0 {
		return core.ErrConflict
	}
	return nil
}
