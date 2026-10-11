package core

import "time"

type ForumThread struct {
	ID         string    `json:"id" bson:"_id"`
	UserID     string    `json:"-" bson:"user_id"`
	Body       string    `json:"body" bson:"body"`
	DistrictID string    `json:"district_id" bson:"district_id"`
	District   string    `json:"district" bson:"district"`
	ProvinceID string    `json:"province_id" bson:"province_id"`
	Province   string    `json:"province" bson:"province"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	CanDelete  bool      `json:"can_delete" bson:"-"`
}

type ForumPage struct {
	Items      []ForumThread `json:"items"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

func BuildForumThread(body, id string, district GeographicDistrict, user User, now time.Time) (ForumThread, error) {
	body, err := ValidateCommunityText(body)
	if err != nil {
		return ForumThread{}, err
	}
	if len(district.ID) != 6 || user.ID == "" {
		return ForumThread{}, ErrInvalid
	}
	return ForumThread{ID: id, UserID: user.ID, Body: body, DistrictID: district.ID, District: district.Name, ProvinceID: district.ID[:4], Province: district.Province, CreatedAt: now}, nil
}

func PublicForumThread(thread ForumThread, user User) ForumThread {
	thread.CanDelete = user.ID != "" && (thread.UserID == user.ID || user.Role == "admin")
	thread.UserID = ""
	return thread
}
