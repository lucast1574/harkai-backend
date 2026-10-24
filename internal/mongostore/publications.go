package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) ListPublications(ctx context.Context, cursor string, limit int) (core.IncidentPage, error) {
	filter := bson.M{"status": bson.M{"$ne": "hidden"}}
	if cursor != "" {
		filter["_id"] = bson.M{"$lt": cursor}
	}
	rows, err := s.DB.Collection("incidents").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(int64(limit+1)))
	if err != nil {
		return core.IncidentPage{}, err
	}
	defer rows.Close(ctx)
	page := core.IncidentPage{Items: []core.Incident{}}
	if err = rows.All(ctx, &page.Items); err != nil {
		return page, err
	}
	if len(page.Items) > limit {
		page.Items = page.Items[:limit]
		page.NextCursor = page.Items[limit-1].ID
	}
	return page, nil
}
