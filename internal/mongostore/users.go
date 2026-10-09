package mongostore

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) CreateUser(ctx context.Context, user core.User) (core.User, error) {
	_, err := s.DB.Collection("users").InsertOne(ctx, user)
	return user, translate(err)
}

func (s *Store) UserByEmail(ctx context.Context, email string) (core.User, error) {
	var user core.User
	err := s.DB.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	return user, translate(err)
}

func (s *Store) UserByID(ctx context.Context, id string) (core.User, error) {
	var user core.User
	err := s.DB.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return user, translate(err)
}

func (s *Store) UpdatePreferences(ctx context.Context, id string, p core.Preferences) (core.User, error) {
	return s.updateUser(ctx, id, bson.M{"preferences": p})
}

func (s *Store) UpdateName(ctx context.Context, id, name string) (core.User, error) {
	return s.updateUser(ctx, id, bson.M{"name": name})
}

func (s *Store) updateUser(ctx context.Context, id string, fields bson.M) (core.User, error) {
	var user core.User
	err := s.DB.Collection("users").FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": fields}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&user)
	return user, translate(err)
}
