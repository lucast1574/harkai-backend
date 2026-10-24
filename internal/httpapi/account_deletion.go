package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

func (a *API) accountDeletionOptions(r *http.Request, u core.User) (Result, error) {
	return Result{Body: map[string]bool{"password": u.PasswordHash != "", "google": u.GoogleSubject != "" && a.Service.VerifyGoogle != nil}}, nil
}
func (a *API) deleteMyAccount(r *http.Request, u core.User) (Result, error) {
	var in struct {
		Confirmation string `json:"confirmation"`
		Password     string `json:"password"`
		GoogleToken  string `json:"google_id_token"`
	}
	if err := decode(r, &in); err != nil {
		return Result{}, err
	}
	err := a.Service.DeleteMyAccount(r.Context(), u, in.Confirmation, in.Password, in.GoogleToken)
	return Result{Status: 202, Body: map[string]string{"status": "deletion_pending"}}, err
}
