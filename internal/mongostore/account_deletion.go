package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

func (s *Store) RequestAccountDeletion(ctx context.Context, id string, now time.Time) error {
	result, err := s.DB.Collection("users").UpdateOne(ctx, bson.M{"_id": id, "deleted_at": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"deleted_at": now, "deletion_pending": true, "deletion_next_at": now.Add(time.Minute), "email": id + "@deleted.invalid", "name": ""}, "$unset": bson.M{"password_hash": "", "google_subject": "", "photo_url": "", "preferences": ""}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return core.ErrNotFound
	}
	return nil
}
func (s *Store) ClaimAccountDeletion(ctx context.Context, now time.Time) (string, error) {
	var record struct {
		ID string `bson:"_id"`
	}
	err := s.DB.Collection("users").FindOneAndUpdate(ctx, bson.M{"deletion_pending": true, "deletion_next_at": bson.M{"$lte": now}}, bson.M{"$set": bson.M{"deletion_next_at": now.Add(2 * time.Minute)}}).Decode(&record)
	return record.ID, translate(err)
}
func (s *Store) AccountDeletionMedia(ctx context.Context, id string) ([]string, error) {
	rows, err := s.DB.Collection("media").Find(ctx, bson.M{"user_id": id})
	if err != nil {
		return nil, err
	}
	defer rows.Close(ctx)
	items := []core.Media{}
	if err = rows.All(ctx, &items); err != nil {
		return nil, err
	}
	ids := []string{}
	for _, m := range items {
		ids = append(ids, m.ID)
	}
	return ids, nil
}
func (s *Store) PurgeAccountRecords(ctx context.Context, id string) error {
	for _, collection := range []string{"sessions", "devices", "business_notices", "comments", "forum_threads", "incidents", "media"} {
		if _, err := s.DB.Collection(collection).DeleteMany(ctx, bson.M{"user_id": id}); err != nil {
			return err
		}
	}
	if _, err := s.DB.Collection("business_points").DeleteMany(ctx, bson.M{"owner_id": id}); err != nil {
		return err
	}
	_, err := s.DB.Collection("users").DeleteOne(ctx, bson.M{"_id": id, "deletion_pending": true})
	return err
}

func (s *Store) ForgetDeletionMedia(ctx context.Context, user, id string) error {
	_, err := s.DB.Collection("media").DeleteOne(ctx, bson.M{"_id": id, "user_id": user})
	return err
}
