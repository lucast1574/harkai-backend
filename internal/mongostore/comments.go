package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func (s *Store) CreateComment(ctx context.Context, c core.Comment) (core.Comment, error) {
	_, err := s.DB.Collection("comments").InsertOne(ctx, c)
	return c, translate(err)
}
func (s *Store) ListComments(ctx context.Context, incident, cursor string) (core.CommentPage, error) {
	filter := bson.M{"incident_id": incident}
	if cursor != "" {
		filter["_id"] = bson.M{"$gt": cursor}
	}
	c, err := s.DB.Collection("comments").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(51))
	if err != nil {
		return core.CommentPage{}, err
	}
	defer c.Close(ctx)
	items := []core.Comment{}
	if err = c.All(ctx, &items); err != nil {
		return core.CommentPage{}, err
	}
	page := core.CommentPage{Items: items}
	if len(items) > 50 {
		page.Items = items[:50]
		page.NextCursor = items[49].ID
	}
	return page, nil
}
func (s *Store) DeleteComment(ctx context.Context, incident, id, owner string, now time.Time) error {
	filter := bson.M{"_id": id, "incident_id": incident, "deleted_at": nil}
	if owner != "" {
		filter["user_id"] = owner
	}
	r, err := s.DB.Collection("comments").UpdateOne(ctx, filter, bson.M{"$set": bson.M{"body": "", "author_name": "", "deleted_at": now}})
	if err != nil {
		return err
	}
	if r.MatchedCount != 1 {
		return core.ErrNotFound
	}
	return nil
}
