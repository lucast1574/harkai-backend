package core

import "time"

type Media struct {
	ID          string    `json:"id" bson:"_id"`
	UserID      string    `json:"-" bson:"user_id"`
	ContentType string    `json:"content_type" bson:"content_type"`
	Size        int       `json:"size" bson:"size"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	Data        []byte    `json:"-" bson:"-"`
}
