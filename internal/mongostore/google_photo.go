package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) UpdateGooglePhoto(ctx context.Context, id, subject, photo string) (core.User, error) {
	var user core.User
	err := s.DB.Collection("users").FindOneAndUpdate(ctx,
		bson.M{"_id": id, "google_subject": subject}, bson.M{"$set": bson.M{"photo_url": photo}},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&user)
	return user, translate(err)
}
