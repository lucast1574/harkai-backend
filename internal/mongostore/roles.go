package mongostore

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"slices"
	"time"
)

func (s *Store) ListUsers(ctx context.Context, after string, limit int) (core.UserPage, error) {
	filter := bson.M{}
	if after != "" {
		filter["_id"] = bson.M{"$lt": after}
	}
	cursor, err := s.DB.Collection("users").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(int64(limit+1)))
	if err != nil {
		return core.UserPage{}, err
	}
	defer cursor.Close(ctx)
	items := []core.User{}
	if err = cursor.All(ctx, &items); err != nil {
		return core.UserPage{}, err
	}
	page := core.UserPage{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		page.NextCursor = page.Items[limit-1].ID
	}
	return page, nil
}

func (s *Store) AssignRole(ctx context.Context, target, actor, role string, now time.Time) (core.User, error) {
	if target == actor || !slices.Contains(core.Roles(), role) {
		return core.User{}, core.ErrForbidden
	}
	// Administrators are protected against demotion here, including concurrent cross-demotions.
	filter := bson.M{"_id": target, "role": bson.M{"$ne": "admin"}}
	update := bson.M{"$set": bson.M{"role": role, "role_assignment": bson.M{"by": actor, "role": role, "at": now}}}
	var user core.User
	err := s.DB.Collection("users").FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		if _, lookup := s.UserByID(ctx, target); lookup != nil {
			return core.User{}, lookup
		}
		return core.User{}, core.ErrForbidden
	}
	return user, translate(err)
}
