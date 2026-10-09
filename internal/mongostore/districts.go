package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

func (s *Store) ArchiveDistricts(ctx context.Context, q core.IncidentQuery, now time.Time) ([]string, error) {
	q.IncludeInactive = true
	q.PublicOnly = true
	q.District = ""
	q.Cursor = ""
	filter := IncidentFilter(q, now)
	filter["district"] = bson.M{"$type": "string", "$ne": ""}
	c, err := s.DB.Collection("incidents").Aggregate(ctx, bson.A{bson.M{"$match": filter}, bson.M{"$group": bson.M{"_id": "$district"}}, bson.M{"$sort": bson.M{"_id": 1}}, bson.M{"$limit": 200}})
	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)
	rows := []struct {
		ID string `bson:"_id"`
	}{}
	if err = c.All(ctx, &rows); err != nil {
		return nil, err
	}
	result := []string{}
	for _, row := range rows {
		result = append(result, row.ID)
	}
	return result, nil
}
