package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func (s *Store) CreateBusiness(ctx context.Context, p core.BusinessPoint) (core.BusinessPoint, error) {
	_, err := s.DB.Collection("business_points").InsertOne(ctx, p)
	return p, translate(err)
}
func (s *Store) BusinessByID(ctx context.Context, id string) (core.BusinessPoint, error) {
	var p core.BusinessPoint
	err := s.DB.Collection("business_points").FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	return p, translate(err)
}
func BusinessPublicFilter(now time.Time) bson.M {
	return bson.M{"enabled": true, "moderation": "approved", "billing_status": "active", "paid_until": bson.M{"$gt": now}}
}
func (s *Store) UpdateBusiness(ctx context.Context, p core.BusinessPoint) (core.BusinessPoint, error) {
	// Update only editable fields. A concurrent payment webhook must survive editing.
	update := bson.M{"name": p.Name, "kind": p.Kind, "subtype": p.Subtype, "description": p.Description, "address": p.Address, "city": p.City, "district": p.District, "latitude": p.Latitude, "longitude": p.Longitude, "location": p.Location, "radius_meters": p.RadiusMeters, "media_id": p.MediaID, "phone": p.Phone, "website": p.Website, "moderation": "pending", "updated_at": p.UpdatedAt}
	var result core.BusinessPoint
	err := s.DB.Collection("business_points").FindOneAndUpdate(ctx, bson.M{"_id": p.ID, "owner_id": p.OwnerID}, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&result)
	return result, translate(err)
}
func (s *Store) ListBusinesses(ctx context.Context, q core.IncidentQuery, owner string, admin bool, now time.Time) (core.BusinessPage, error) {
	filter := BusinessPublicFilter(now)
	if owner != "" || admin {
		filter = bson.M{}
		if owner != "" {
			filter["owner_id"] = owner
		}
	} else {
		filter["location"] = bson.M{"$geoWithin": bson.M{"$centerSphere": bson.A{bson.A{q.Longitude, q.Latitude}, float64(q.RadiusMeters) / 6371008.8}}}
	}
	if q.Cursor != "" {
		filter["_id"] = bson.M{"$lt": q.Cursor}
	}
	rows, err := s.DB.Collection("business_points").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(int64(q.Limit+1)))
	if err != nil {
		return core.BusinessPage{}, err
	}
	defer rows.Close(ctx)
	page := core.BusinessPage{Items: []core.BusinessPoint{}}
	if err = rows.All(ctx, &page.Items); err != nil {
		return page, err
	}
	if len(page.Items) > q.Limit {
		page.Items = page.Items[:q.Limit]
		page.NextCursor = page.Items[q.Limit-1].ID
	}
	return page, nil
}
func (s *Store) ModerateBusiness(ctx context.Context, id, status string, now time.Time) (core.BusinessPoint, error) {
	var p core.BusinessPoint
	err := s.DB.Collection("business_points").FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"moderation": status, "updated_at": now}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&p)
	return p, translate(err)
}
func (s *Store) SetBusinessEnabled(ctx context.Context, id, owner string, enabled bool, now time.Time) (core.BusinessPoint, error) {
	var p core.BusinessPoint
	err := s.DB.Collection("business_points").FindOneAndUpdate(ctx, bson.M{"_id": id, "owner_id": owner}, bson.M{"$set": bson.M{"enabled": enabled, "updated_at": now}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&p)
	return p, translate(err)
}
func (s *Store) ApplyBusinessEntitlement(ctx context.Context, e core.BusinessEntitlement, now time.Time) error {
	filter := bson.M{"_id": e.PointID, "owner_id": e.OwnerID, "$or": bson.A{bson.M{"billing_event_at": bson.M{"$lt": e.EventAt}}, bson.M{"billing_event_at": bson.M{"$exists": false}}, bson.M{"billing_event_id": e.EventID, "billing_event_at": e.EventAt}}}
	update := bson.M{"billing_status": e.Status, "paid_until": e.PaidUntil, "billing_event_at": e.EventAt, "billing_event_id": e.EventID, "subscription_id": e.SubscriptionID, "updated_at": now}
	result, err := s.DB.Collection("business_points").UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return core.ErrConflict
	}
	return nil
}
