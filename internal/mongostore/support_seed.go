package mongostore

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

//go:embed support-seed/directory.json
var supportDirectorySeed []byte

//go:embed support-seed/places.json
var supportPlacesSeed []byte

//go:embed support-seed/renipress.json
var supportRegistrySeed []byte

// SeedSupport inserts missing curated records; reruns never overwrite edited records.
func (s *Store) SeedSupport(ctx context.Context) error {
	districts, places, err := ReadSupportSeed()
	if err != nil {
		return err
	}
	models := make([]mongo.WriteModel, 0, len(districts))
	for _, d := range districts {
		models = append(models, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": d.ID}).SetUpdate(bson.M{"$setOnInsert": d}).SetUpsert(true))
	}
	if _, err = s.DB.Collection("support_directory").BulkWrite(ctx, models); err != nil {
		return err
	}
	models = make([]mongo.WriteModel, 0, len(places))
	for _, p := range places {
		models = append(models, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": p.ID}).SetUpdate(bson.M{"$setOnInsert": p}).SetUpsert(true))
	}
	if len(models) > 0 {
		if _, err = s.DB.Collection("support_places").BulkWrite(ctx, models); err != nil {
			return err
		}
	}
	for _, city := range []string{"lima", "trujillo"} {
		c := core.EmergencyContacts{ID: "PE:" + city, Country: "PE", City: city, Numbers: map[string]string{"police": "105", "firefighters": "116", "medical": "106"}}
		if _, err = s.DB.Collection("emergency_contacts").UpdateOne(ctx, bson.M{"_id": c.ID}, bson.M{"$setOnInsert": c}, options.UpdateOne().SetUpsert(true)); err != nil {
			return err
		}
	}
	return nil
}
func ReadSupportSeed() ([]core.SupportDistrict, []core.SupportPlace, error) {
	var districts []core.SupportDistrict
	var places []core.SupportPlace
	if err := json.Unmarshal(supportDirectorySeed, &districts); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(supportPlacesSeed, &places); err != nil {
		return nil, nil, err
	}
	var registry []core.SupportPlace
	if err := json.Unmarshal(supportRegistrySeed, &registry); err != nil {
		return nil, nil, err
	}
	places = append(places, registry...)
	var err error
	districts, err = nationalSupportDirectory(districts)
	if err != nil {
		return nil, nil, err
	}
	for n := range districts {
		hash := sha256.Sum256([]byte(districts[n].City + ":" + districts[n].District))
		districts[n].ID = hex.EncodeToString(hash[:])
		if err := ValidateSupportDistrict(districts[n]); err != nil {
			return nil, nil, err
		}
	}
	for n := range places {
		places[n].Location = core.Point{Type: "Point", Coordinates: []float64{places[n].Longitude, places[n].Latitude}}
		if err := ValidateSupportPlace(places[n]); err != nil {
			return nil, nil, err
		}
	}
	return districts, places, nil
}
