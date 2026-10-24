package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net"
	"net/http"
	"time"
)

func (a *API) authLimit(r *http.Request, key string, max int, window time.Duration) error {
	return a.Service.Limits.Allow(r.Context(), key, max, window, a.Service.Now())
}

func remoteKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (a *API) register(r *http.Request, _ core.User) (Result, error) {
	if err := a.authLimit(r, "register:"+remoteKey(r), 10, time.Hour); err != nil {
		return Result{}, err
	}
	var input core.Credentials
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	result, err := a.Service.Register(r.Context(), input)
	return Result{Status: 201, Body: result}, err
}

func (a *API) login(r *http.Request, _ core.User) (Result, error) {
	if err := a.authLimit(r, "login-ip:"+remoteKey(r), 60, time.Minute); err != nil {
		return Result{}, err
	}
	var input core.Credentials
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	email, err := core.NormalizeEmail(input.Email)
	if err != nil {
		return Result{}, core.ErrUnauthorized
	}
	if err = a.authLimit(r, "login-account:"+a.Service.HashToken(email), 10, 15*time.Minute); err != nil {
		return Result{}, err
	}
	result, err := a.Service.Login(r.Context(), input)
	return Result{Body: result}, err
}

func (a *API) refresh(r *http.Request, _ core.User) (Result, error) {
	if err := a.authLimit(r, "refresh:"+remoteKey(r), 120, time.Minute); err != nil {
		return Result{}, err
	}
	var input struct {
		RefreshToken string `json:"refresh_token"`
		RotationID   string `json:"rotation_id,omitempty"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	result, err := a.Service.RefreshWithRotation(r.Context(), input.RefreshToken, input.RotationID)
	return Result{Body: result}, err
}

func (a *API) logout(r *http.Request, _ core.User) (Result, error) {
	token := r.Header.Get("Authorization")[len("Bearer "):]
	return Result{Status: 204}, a.Service.Logout(r.Context(), token)
}
