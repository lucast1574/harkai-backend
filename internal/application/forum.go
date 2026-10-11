package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

type ForumStore interface {
	CreateForumThread(context.Context, core.ForumThread) (core.ForumThread, error)
	ForumThreadByID(context.Context, string) (core.ForumThread, error)
	ListForumThreads(context.Context, string, string, string) (core.ForumPage, error)
	DeleteForumThread(context.Context, string, string, time.Time) error
}

func (s Service) ForumFeed(ctx context.Context, user core.User, districtID, scope, cursor string) (core.ForumPage, error) {
	if scope != "district" && scope != "city" || cursor != "" && !core.ValidID(cursor) {
		return core.ForumPage{}, core.ErrInvalid
	}
	district, err := s.DistrictByID(districtID)
	if err != nil {
		return core.ForumPage{}, err
	}
	districtFilter := district.ID
	if scope == "city" {
		districtFilter = ""
	}
	page, err := s.Forum.ListForumThreads(ctx, district.ID[:4], districtFilter, cursor)
	for n := range page.Items {
		page.Items[n] = core.PublicForumThread(page.Items[n], user)
	}
	return page, err
}

func (s Service) StartForumThread(ctx context.Context, user core.User, districtID, body string) (core.ForumThread, error) {
	district, err := s.DistrictByID(districtID)
	if err != nil {
		return core.ForumThread{}, err
	}
	thread, err := core.BuildForumThread(body, s.ID(), district, user, s.Now())
	if err != nil {
		return thread, err
	}
	if err = s.Limits.Allow(ctx, "forum-hour:"+user.ID, 10, time.Hour, s.Now()); err != nil {
		return thread, err
	}
	thread, err = s.Forum.CreateForumThread(ctx, thread)
	return core.PublicForumThread(thread, user), err
}

func (s Service) RemoveForumThread(ctx context.Context, user core.User, id string) error {
	if !core.ValidID(id) {
		return core.ErrInvalid
	}
	owner := user.ID
	if user.Role == "admin" {
		owner = ""
	}
	return s.Forum.DeleteForumThread(ctx, id, owner, s.Now())
}

func (s Service) ForumComments(ctx context.Context, user core.User, id, cursor string) (core.CommentPage, error) {
	if !core.ValidID(id) || cursor != "" && !core.ValidID(cursor) {
		return core.CommentPage{}, core.ErrInvalid
	}
	if _, err := s.Forum.ForumThreadByID(ctx, id); err != nil {
		return core.CommentPage{}, err
	}
	page, err := s.Comments.ListComments(ctx, id, cursor)
	for n := range page.Items {
		page.Items[n] = core.PublicComment(page.Items[n], user)
	}
	return page, err
}

func (s Service) ReplyForumThread(ctx context.Context, user core.User, id, body string) (core.Comment, error) {
	comment, err := core.BuildComment(body, s.ID(), id, user, s.Now())
	if err != nil {
		return comment, err
	}
	if _, err = s.Forum.ForumThreadByID(ctx, id); err != nil {
		return comment, err
	}
	if err = s.Limits.Allow(ctx, "comment-minute:"+user.ID, 5, time.Minute, s.Now()); err != nil {
		return comment, err
	}
	if err = s.Limits.Allow(ctx, "comment-hour:"+user.ID, 30, time.Hour, s.Now()); err != nil {
		return comment, err
	}
	comment, err = s.Comments.CreateComment(ctx, comment)
	return core.PublicComment(comment, user), err
}

func (s Service) RemoveForumReply(ctx context.Context, user core.User, id, commentID string) error {
	if !core.ValidID(id) || !core.ValidID(commentID) {
		return core.ErrInvalid
	}
	if _, err := s.Forum.ForumThreadByID(ctx, id); err != nil {
		return err
	}
	owner := user.ID
	if user.Role == "admin" {
		owner = ""
	}
	return s.Comments.DeleteComment(ctx, id, commentID, owner, s.Now())
}
