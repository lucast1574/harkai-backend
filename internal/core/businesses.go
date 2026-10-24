package core

import "time"

type BusinessPoint struct {
	Offer          string     `json:"offer,omitempty" bson:"offer,omitempty"`
	DiscountCode   string     `json:"discount_code,omitempty" bson:"discount_code,omitempty"`
	PaymentMode    string     `json:"payment_mode" bson:"payment_mode"`
	PaidAt         *time.Time `json:"paid_at,omitempty" bson:"paid_at,omitempty"`
	PaymentID      string     `json:"-" bson:"payment_id,omitempty"`
	ID             string     `json:"id" bson:"_id"`
	OwnerID        string     `json:"-" bson:"owner_id"`
	Name           string     `json:"name" bson:"name"`
	Kind           string     `json:"kind" bson:"kind"`
	Subtype        string     `json:"subtype,omitempty" bson:"subtype,omitempty"`
	Description    string     `json:"description" bson:"description"`
	Address        string     `json:"address" bson:"address"`
	City           string     `json:"city" bson:"city"`
	District       string     `json:"district" bson:"district"`
	Latitude       float64    `json:"latitude" bson:"latitude"`
	Longitude      float64    `json:"longitude" bson:"longitude"`
	Location       Point      `json:"-" bson:"location"`
	RadiusMeters   int        `json:"radius_meters" bson:"radius_meters"`
	MediaID        string     `json:"media_id,omitempty" bson:"media_id,omitempty"`
	Phone          string     `json:"phone,omitempty" bson:"phone,omitempty"`
	Website        string     `json:"website,omitempty" bson:"website,omitempty"`
	Moderation     string     `json:"moderation" bson:"moderation"`
	BillingStatus  string     `json:"billing_status" bson:"billing_status"`
	PaidUntil      time.Time  `json:"paid_until,omitempty" bson:"paid_until"`
	BillingEventAt time.Time  `json:"-" bson:"billing_event_at"`
	BillingEventID string     `json:"-" bson:"billing_event_id"`
	SubscriptionID string     `json:"-" bson:"subscription_id"`
	Enabled        bool       `json:"enabled" bson:"enabled"`
	Sponsored      bool       `json:"sponsored" bson:"-"`
	CreatedAt      time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" bson:"updated_at"`
}
type BusinessInput struct {
	Offer        string  `json:"offer,omitempty"`
	DiscountCode string  `json:"discount_code,omitempty"`
	Name         string  `json:"name"`
	Kind         string  `json:"kind"`
	Subtype      string  `json:"subtype,omitempty"`
	Description  string  `json:"description"`
	Address      string  `json:"address"`
	City         string  `json:"city"`
	District     string  `json:"district"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeters int     `json:"radius_meters"`
	MediaID      string  `json:"media_id,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	Website      string  `json:"website,omitempty"`
}
type BusinessEntitlement struct {
	PointID, OwnerID, EventID, SubscriptionID, Status string
	PaymentID, Mode                                   string
	PaidAt                                            *time.Time
	EventAt, PaidUntil                                time.Time
}
type BusinessPage struct {
	Items      []BusinessPoint `json:"items"`
	NextCursor string          `json:"next_cursor,omitempty"`
}
type BusinessNotice struct {
	Kind           string    `json:"kind,omitempty" bson:"kind,omitempty"`
	Stage          int       `json:"stage,omitempty" bson:"stage,omitempty"`
	Title          string    `json:"title,omitempty" bson:"title,omitempty"`
	Body           string    `json:"body,omitempty" bson:"body,omitempty"`
	ID             string    `json:"id" bson:"_id"`
	UserID         string    `json:"-" bson:"user_id"`
	DeviceID       string    `json:"-" bson:"device_id"`
	SessionID      string    `json:"-" bson:"session_id"`
	PointID        string    `json:"point_id" bson:"point_id"`
	Name           string    `json:"name" bson:"name"`
	Pending        bool      `json:"-" bson:"pending"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	NextEligibleAt time.Time `json:"-" bson:"next_eligible_at"`
}

func (b BusinessPoint) Public(now time.Time) bool {
	return b.Enabled && b.Moderation == "approved" && b.BillingStatus == "active" && ((b.PaymentMode == "one_time" && b.PaidAt != nil && !b.PaidAt.After(now)) || (b.PaymentMode != "one_time" && b.PaidUntil.After(now)))
}
func PublicBusiness(b BusinessPoint) BusinessPoint { b.Sponsored = true; return b }
func BusinessKinds() []string                      { return []string{"restaurant", "health", "shop", "service", "other"} }
