package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

func (s *Store) CreateMedia(ctx context.Context, media core.Media) (core.Media, error) {
	_, err := s.DB.Collection("media").InsertOne(ctx, media)
	return media, translate(err)
}

func (s *Store) MediaByID(ctx context.Context, id string) (core.Media, error) {
	var media core.Media
	err := s.DB.Collection("media").FindOne(ctx, bson.M{"_id": id}).Decode(&media)
	return media, translate(err)
}

func (s *Store) MediaIsPublic(ctx context.Context, id string, now time.Time) (bool, error) {
	filter := IncidentFilter(core.IncidentQuery{OwnerID: "unused", IncludeInactive: true, PublicOnly: true}, now)
	delete(filter, "user_id")
	filter["media_id"] = id
	count, err := s.DB.Collection("incidents").CountDocuments(ctx, filter)
	return count > 0, err
}
