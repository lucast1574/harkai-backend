package mongostore

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

func (s *Store) CreateSession(ctx context.Context, session core.Session) error {
	_, err := s.DB.Collection("sessions").InsertOne(ctx, session)
	return translate(err)
}

func (s *Store) SessionByAccess(ctx context.Context, hash string, now time.Time) (core.Session, error) {
	var session core.Session
	err := s.DB.Collection("sessions").FindOne(ctx, bson.M{"access_hash": hash, "access_expires_at": bson.M{"$gt": now}, "expires_at": bson.M{"$gt": now}}).Decode(&session)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return session, core.ErrUnauthorized
	}
	return session, err
}

func (s *Store) RotateSession(ctx context.Context, hash string, next core.Session, now time.Time) (core.Session, error) {
	var previous core.Session
	// Conditional single-document rotation: exactly one request can consume a refresh token.
	err := s.DB.Collection("sessions").FindOneAndUpdate(ctx, bson.M{"refresh_hash": hash, "expires_at": bson.M{"$gt": now}}, bson.M{"$set": bson.M{"access_hash": next.AccessHash, "refresh_hash": next.RefreshHash, "access_expires_at": next.AccessExpiresAt}}).Decode(&previous)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return previous, core.ErrUnauthorized
	}
	return previous, err
}

func (s *Store) DeleteSession(ctx context.Context, hash string) error {
	_, err := s.DB.Collection("sessions").DeleteOne(ctx, bson.M{"access_hash": hash})
	return err
}
