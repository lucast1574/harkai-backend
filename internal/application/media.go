package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

type MediaStore interface {
	CreateMedia(context.Context, core.Media) (core.Media, error)
	MediaByID(context.Context, string) (core.Media, error)
	MediaIsPublic(context.Context, string, time.Time) (bool, error)
}

type PhotoStore interface {
	Put(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, error)
}

func (s Service) Upload(ctx context.Context, userID string, raw []byte) (core.Media, error) {
	if err := s.Limits.Allow(ctx, "upload:"+userID, 20, time.Hour, s.Now()); err != nil {
		return core.Media{}, err
	}
	data, err := s.NormalizeImage(raw)
	if err != nil {
		return core.Media{}, err
	}
	media := core.Media{ID: s.ID(), UserID: userID, ContentType: "image/jpeg", Size: len(data), CreatedAt: s.Now()}
	if err = s.Photos.Put(ctx, media.ID, data); err != nil {
		return core.Media{}, err
	}
	return s.Media.CreateMedia(ctx, media)
}

func (s Service) Download(ctx context.Context, id string, user core.User) (core.Media, error) {
	media, err := s.Media.MediaByID(ctx, id)
	if err != nil {
		return core.Media{}, err
	}
	if media.UserID == user.ID || user.Role == "admin" {
		media.Data, err = s.Photos.Get(ctx, media.ID)
		return media, err
	}
	public, err := s.Media.MediaIsPublic(ctx, id, s.Now())
	if err != nil {
		return core.Media{}, err
	}
	if !public {
		return core.Media{}, core.ErrNotFound
	}
	media.Data, err = s.Photos.Get(ctx, media.ID)
	return media, err
}

func (s Service) ValidateMedia(ctx context.Context, userID, id string) error {
	if id == "" {
		return nil
	}
	media, err := s.Media.MediaByID(ctx, id)
	if err != nil {
		return err
	}
	if media.UserID != userID {
		return core.ErrForbidden
	}
	return nil
}
