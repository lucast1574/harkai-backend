package core

import "time"

type Session struct {
	ID              string    `bson:"_id"`
	UserID          string    `bson:"user_id"`
	AccessHash      string    `bson:"access_hash"`
	RefreshHash     string    `bson:"refresh_hash"`
	AccessExpiresAt time.Time `bson:"access_expires_at"`
	ExpiresAt       time.Time `bson:"expires_at"`
}

type AuthResult struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}
