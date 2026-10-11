package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) SupportDirectory(ctx context.Context, city string) ([]core.SupportDistrict, error) {
	cursor, err := s.DB.Collection("support_directory").Find(ctx, bson.M{"city": city}, options.Find().SetSort(bson.D{{Key: "district", Value: 1}}).SetLimit(100))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	items := make([]core.SupportDistrict, 0)
	err = cursor.All(ctx, &items)
	return items, err
}
func (s *Store) SupportPlaces(ctx context.Context, q core.IncidentQuery) ([]core.SupportPlace, error) {
	filter := bson.M{"location": bson.M{"$near": bson.M{"$geometry": core.Point{Type: "Point", Coordinates: []float64{q.Longitude, q.Latitude}}, "$maxDistance": q.RadiusMeters}}}
	if q.Boundary != nil {
		filter["location"] = bson.M{"$geoWithin": bson.M{"$geometry": q.Boundary}}
	}
	cursor, err := s.DB.Collection("support_places").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	items := make([]core.SupportPlace, 0)
	err = cursor.All(ctx, &items)
	return items, err
}
