package core

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type Comment struct {
	ID         string     `json:"id" bson:"_id"`
	IncidentID string     `json:"incident_id" bson:"incident_id"`
	UserID     string     `json:"-" bson:"user_id"`
	AuthorName string     `json:"author_name" bson:"author_name"`
	Body       string     `json:"body" bson:"body"`
	CreatedAt  time.Time  `json:"created_at" bson:"created_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	CanDelete  bool       `json:"can_delete" bson:"-"`
}
type CommentPage struct {
	Items      []Comment `json:"items"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

func BuildComment(body, id, incidentID string, u User, now time.Time) (Comment, error) {
	body = strings.TrimSpace(body)
	if utf8.RuneCountInString(body) < 2 || utf8.RuneCountInString(body) > 1000 || strings.ContainsRune(body, '\x00') {
		return Comment{}, fmt.Errorf("%w: comment must contain 2 to 1000 characters", ErrInvalid)
	}
	if !ValidID(incidentID) {
		return Comment{}, ErrInvalid
	}
	return Comment{ID: id, IncidentID: incidentID, UserID: u.ID, AuthorName: u.Name, Body: body, CreatedAt: now}, nil
}
func ValidID(id string) bool { return len(id) == 24 && strings.Trim(id, "0123456789abcdef") == "" }
func PublicComment(c Comment, u User) Comment {
	c.CanDelete = u.ID != "" && (u.ID == c.UserID || u.Role == "admin") && c.DeletedAt == nil
	if c.DeletedAt != nil {
		c.Body = ""
		c.AuthorName = ""
	}
	return c
}
