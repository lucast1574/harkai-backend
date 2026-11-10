package core

import "time"

// A checkout attempt is persisted before calling a provider so retries on any
// replica share the same idempotency key. Only an expired session is replaced.
type BusinessCheckout struct {
	PointID   string    `bson:"_id"`
	OwnerID   string    `bson:"owner_id"`
	Key       string    `bson:"key"`
	SessionID string    `bson:"session_id"`
	CreatedAt time.Time `bson:"created_at"`
}
type CheckoutSession struct {
	ID     string `json:"-"`
	URL    string `json:"url"`
	Status string `json:"status"`
}
