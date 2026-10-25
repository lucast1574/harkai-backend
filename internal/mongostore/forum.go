package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func (s *Store) CreateForumThread(ctx context.Context, thread core.ForumThread) (core.ForumThread, error) {
	_, err := s.DB.Collection("forum_threads").InsertOne(ctx, thread)
	return thread, translate(err)
}
func (s *Store) ForumThreadByID(ctx context.Context, id string) (core.ForumThread, error) {
	var thread core.ForumThread
	err := s.DB.Collection("forum_threads").FindOne(ctx, bson.M{"_id": id, "deleted_at": nil}).Decode(&thread)
	return thread, translate(err)
}
func (s *Store) ListForumThreads(ctx context.Context, province, district, cursor string) (core.ForumPage, error) {
	filter := bson.M{"province_id": province, "deleted_at": nil}
	if len(province) == 2 {
		filter["province_id"] = bson.M{"$regex": "^" + province}
	}
	if province == "metro-lima" {
		filter["province_id"] = bson.M{"$in": []string{"1501", "0701"}}
	}
	if district != "" {
		filter["district_id"] = district
	}
	if cursor != "" {
		filter["_id"] = bson.M{"$lt": cursor}
	}
	rows, err := s.DB.Collection("forum_threads").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(31))
	if err != nil {
		return core.ForumPage{}, err
	}
	defer rows.Close(ctx)
	items := []core.ForumThread{}
	if err = rows.All(ctx, &items); err != nil {
		return core.ForumPage{}, err
	}
	page := core.ForumPage{Items: items}
	if len(items) > 30 {
		page.Items = items[:30]
		page.NextCursor = items[29].ID
	}
	return page, nil
}
func (s *Store) DeleteForumThread(ctx context.Context, id, owner string, now time.Time) error {
	filter := bson.M{"_id": id, "deleted_at": nil}
	if owner != "" {
		filter["user_id"] = owner
	}
	result, err := s.DB.Collection("forum_threads").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"deleted_at": now, "body": ""}})
	if err != nil {
		return err
	}
	if result.MatchedCount != 1 {
		return core.ErrNotFound
	}
	return nil
}
