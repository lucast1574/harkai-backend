package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

func (a *API) forum(r *http.Request, user core.User) (Result, error) {
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "district"
	}
	page, err := a.Service.ForumFeed(r.Context(), user, r.URL.Query().Get("ubigeo"), scope, r.URL.Query().Get("cursor"))
	return Result{Body: page}, err
}
func (a *API) startForum(r *http.Request, user core.User) (Result, error) {
	var input struct {
		Body       string `json:"body"`
		DistrictID string `json:"district_id"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	thread, err := a.Service.StartForumThread(r.Context(), user, input.DistrictID, input.Body)
	return Result{Status: 201, Body: thread}, err
}
func (a *API) deleteForum(r *http.Request, user core.User) (Result, error) {
	return Result{Status: 204}, a.Service.RemoveForumThread(r.Context(), user, r.PathValue("id"))
}
func (a *API) forumComments(r *http.Request, user core.User) (Result, error) {
	page, err := a.Service.ForumComments(r.Context(), user, r.PathValue("id"), r.URL.Query().Get("cursor"))
	return Result{Body: page}, err
}
func (a *API) forumReply(r *http.Request, user core.User) (Result, error) {
	var input struct {
		Body string `json:"body"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	comment, err := a.Service.ReplyForumThread(r.Context(), user, r.PathValue("id"), input.Body)
	return Result{Status: 201, Body: comment}, err
}
func (a *API) deleteForumReply(r *http.Request, user core.User) (Result, error) {
	return Result{Status: 204}, a.Service.RemoveForumReply(r.Context(), user, r.PathValue("id"), r.PathValue("commentID"))
}
