package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

type Comments interface {
	CreateComment(context.Context, core.Comment) (core.Comment, error)
	ListComments(context.Context, string, string) (core.CommentPage, error)
	DeleteComment(context.Context, string, string, string, time.Time) error
}

func (s Service) CommentThread(ctx context.Context, u core.User, id, cursor string) (core.CommentPage, error) {
	if !core.ValidID(id) || (cursor != "" && !core.ValidID(cursor)) {
		return core.CommentPage{}, core.ErrInvalid
	}
	if _, err := s.Incident(ctx, id, u); err != nil {
		return core.CommentPage{}, err
	}
	page, err := s.Comments.ListComments(ctx, id, cursor)
	for n := range page.Items {
		page.Items[n] = core.PublicComment(page.Items[n], u)
	}
	return page, err
}
func (s Service) AddComment(ctx context.Context, u core.User, id, body string) (core.Comment, error) {
	comment, err := core.BuildComment(body, s.ID(), id, u, s.Now())
	if err != nil {
		return comment, err
	}
	if _, err = s.Incident(ctx, id, u); err != nil {
		return comment, err
	}
	if err = s.Limits.Allow(ctx, "comment-minute:"+u.ID, 5, time.Minute, s.Now()); err != nil {
		return comment, err
	}
	if err = s.Limits.Allow(ctx, "comment-hour:"+u.ID, 30, time.Hour, s.Now()); err != nil {
		return comment, err
	}
	comment, err = s.Comments.CreateComment(ctx, comment)
	return core.PublicComment(comment, u), err
}
func (s Service) RemoveComment(ctx context.Context, u core.User, incidentID, id string) error {
	if !core.ValidID(id) || !core.ValidID(incidentID) {
		return core.ErrInvalid
	}
	if _, err := s.Incident(ctx, incidentID, u); err != nil {
		return err
	}
	owner := u.ID
	if u.Role == "admin" {
		owner = ""
	}
	return s.Comments.DeleteComment(ctx, incidentID, id, owner, s.Now())
}
