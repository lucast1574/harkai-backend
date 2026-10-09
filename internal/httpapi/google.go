package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
	"time"
)

func (a *API) googleLogin(r *http.Request, _ core.User) (Result, error) {
	if err := a.authLimit(r, "google:"+remoteKey(r), 30, time.Minute); err != nil {
		return Result{}, err
	}
	var input struct {
		IDToken string `json:"id_token"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	result, err := a.Service.GoogleLogin(r.Context(), input.IDToken)
	return Result{Body: result}, err
}
