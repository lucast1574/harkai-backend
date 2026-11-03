package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (s *Store) ListForumDistricts(ctx context.Context) ([]core.ForumDistrict, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		{{Key: "$group", Value: bson.M{
			"_id": "$district_id", "district": bson.M{"$first": "$district"},
			"province": bson.M{"$first": "$province"}, "messages": bson.M{"$sum": 1},
			"updated_at": bson.M{"$max": "$created_at"},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "updated_at", Value: -1}, {Key: "_id", Value: 1}}}},
	}
	rows, err := s.DB.Collection("forum_threads").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close(ctx)
	items := []core.ForumDistrict{}
	if err = rows.All(ctx, &items); err != nil {
		return nil, translate(err)
	}
	return items, nil
}
