package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func (s *Store) CreateIncident(ctx context.Context, i core.Incident) (core.Incident, error) {
	_, err := s.DB.Collection("incidents").InsertOne(ctx, i)
	return i, translate(err)
}

func (s *Store) IncidentByID(ctx context.Context, id string) (core.Incident, error) {
	var i core.Incident
	err := s.DB.Collection("incidents").FindOne(ctx, bson.M{"_id": id}).Decode(&i)
	return i, translate(err)
}

func (s *Store) ListIncidents(ctx context.Context, q core.IncidentQuery, now time.Time) (core.IncidentPage, error) {
	filter := IncidentFilter(q, now)
	cursor, err := s.DB.Collection("incidents").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(int64(q.Limit+1)))
	if err != nil {
		return core.IncidentPage{}, err
	}
	defer cursor.Close(ctx)
	items := []core.Incident{}
	if err = cursor.All(ctx, &items); err != nil {
		return core.IncidentPage{}, err
	}
	page := core.IncidentPage{Items: items}
	if len(items) > q.Limit {
		page.Items = items[:q.Limit]
		page.NextCursor = page.Items[q.Limit-1].ID
	}
	return page, nil
}

func IncidentFilter(q core.IncidentQuery, now time.Time) bson.M {
	filter := bson.M{"type": bson.M{"$in": core.IncidentTypes()}}
	if !q.IncludeInactive {
		filter["status"] = "active"
		filter["$or"] = bson.A{bson.M{"expires_at": bson.M{"$exists": false}}, bson.M{"expires_at": nil}, bson.M{"expires_at": bson.M{"$gt": now}}}
	}
	if q.PublicOnly && q.IncludeInactive {
		filter["status"] = bson.M{"$in": bson.A{"active", "resolved"}}
	}
	if q.District != "" {
		filter["district"] = q.District
	}
	if q.Boundary != nil {
		filter["location"] = bson.M{"$geoWithin": bson.M{"$geometry": q.Boundary}}
	}
	if q.OwnerID != "" {
		filter["user_id"] = q.OwnerID
	} else if q.Boundary == nil {
		filter["location"] = bson.M{"$geoWithin": bson.M{"$centerSphere": bson.A{bson.A{q.Longitude, q.Latitude}, float64(q.RadiusMeters) / 6371008.8}}}
	}
	if q.Type != "" {
		filter["type"] = q.Type
		if q.Type == "crime" {
			filter["type"] = bson.M{"$in": core.CrimeTypes()}
		}
	}
	if q.Cursor != "" {
		filter["_id"] = bson.M{"$lt": q.Cursor}
	}
	dateFilter := bson.M{}
	if !q.IncludeInactive {
		dateFilter["$gt"] = now.Add(-3 * time.Hour)
	}
	if q.After != nil {
		dateFilter["$gte"] = *q.After
	}
	if q.Before != nil {
		dateFilter["$lt"] = *q.Before
	}
	if len(dateFilter) > 0 {
		filter["created_at"] = dateFilter
	}
	return filter
}

func (s *Store) SetIncidentStatus(ctx context.Context, id, owner, status string, now time.Time) (core.Incident, error) {
	filter := bson.M{"_id": id}
	if owner != "" {
		filter["user_id"] = owner
	}
	var i core.Incident
	err := s.DB.Collection("incidents").FindOneAndUpdate(ctx, filter, bson.M{"$set": bson.M{"status": status, "updated_at": now}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&i)
	return i, translate(err)
}
