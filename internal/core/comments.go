package core

import "time"

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
	body, err := ValidateCommunityText(body)
	if err != nil {
		return Comment{}, err
	}
	if !ValidID(incidentID) {
		return Comment{}, ErrInvalid
	}
	return Comment{ID: id, IncidentID: incidentID, UserID: u.ID, AuthorName: "Vecino anónimo", Body: body, CreatedAt: now}, nil
}
func ValidID(id string) bool {
	if len(id) != 24 {
		return false
	}
	for _, r := range id {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
func PublicComment(c Comment, u User) Comment {
	c.CanDelete = u.ID != "" && (u.ID == c.UserID || u.Role == "admin") && c.DeletedAt == nil
	c.AuthorName = "Vecino anónimo"
	if !CommunityTextAllowed(c.Body) {
		c.Body = "Comentario oculto por lenguaje inapropiado."
	}
	if c.DeletedAt != nil {
		c.Body = ""
		c.AuthorName = ""
	}
	return c
}
