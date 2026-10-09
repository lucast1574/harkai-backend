package mongostore

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Store struct{ DB *mongo.Database }

func Connect(ctx context.Context, uri, database string) (*Store, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetMaxPoolSize(30).SetServerSelectionTimeout(5 * time.Second))
	if err != nil {
		return nil, errors.New("database configuration failed")
	}
	if err = client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, errors.New("database connection failed")
	}
	return &Store{DB: client.Database(database)}, nil
}

func (s *Store) Close(ctx context.Context) error { return s.DB.Client().Disconnect(ctx) }
func (s *Store) Ping(ctx context.Context) error  { return s.DB.Client().Ping(ctx, nil) }
func NewID() string                              { return bson.NewObjectID().Hex() }

func translate(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return core.ErrNotFound
	}
	if mongo.IsDuplicateKeyError(err) {
		return core.ErrConflict
	}
	return err
}
