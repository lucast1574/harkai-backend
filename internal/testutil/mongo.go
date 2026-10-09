package testutil

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"os"
	"testing"
	"time"
)

func Store(t *testing.T) *mongostore.Store {
	t.Helper()
	uri := os.Getenv("TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("TEST_MONGODB_URI required for isolated MongoDB integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	store, err := mongostore.Connect(ctx, uri, "harkai_test_"+mongostore.NewID())
	if err != nil {
		t.Fatal(err)
	}
	if err = store.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := store.DB.Drop(ctx); err != nil {
			t.Errorf("drop test database: %v", err)
		}
		if err := store.Close(ctx); err != nil {
			t.Errorf("disconnect: %v", err)
		}
	})
	return store
}
