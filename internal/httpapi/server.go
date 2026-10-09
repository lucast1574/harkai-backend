package httpapi

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Result struct {
	Status  int
	Body    any
	Headers http.Header
	Bytes   []byte
}
type Action func(*http.Request, core.User) (Result, error)
type API struct {
	Service application.Service
	Ping    func(context.Context) error
	Origins []string
	Logger  *slog.Logger
	router  *http.ServeMux
}
type endpoint struct {
	api       *API
	action    Action
	protected bool
}

func New(s application.Service, ping func(context.Context) error, origins []string, logger *slog.Logger) *API {
	a := &API{Service: s, Ping: ping, Origins: origins, Logger: logger, router: http.NewServeMux()}
	_ = a.routes()
	return a
}

// The net/http interface mandates a method with no return value. All actions return Result,error.
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer a.recoverRequest(w, r)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "no-store")
	origin := r.Header.Get("Origin")
	if origin != "" {
		if !slices.Contains(a.Origins, origin) {
			_ = writeError(w, core.ErrForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	timeout := 10 * time.Second
	if r.URL.Path == "/v1/analysis/audio" || r.URL.Path == "/v1/media" {
		timeout = 40 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	a.router.ServeHTTP(w, r.WithContext(ctx))
}

func (e endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var user core.User
	auth := r.Header.Get("Authorization")
	if auth != "" || e.protected {
		if !strings.HasPrefix(auth, "Bearer ") {
			_ = writeError(w, core.ErrUnauthorized)
			return
		}
		var err error
		user, err = e.api.Service.Authenticate(r.Context(), strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			_ = writeError(w, err)
			return
		}
	}
	result, err := e.action(r, user)
	if err != nil {
		_ = writeError(w, err)
		return
	}
	if err = writeResult(w, result); err != nil {
		e.api.Logger.Error("response write failed", "method", r.Method, "error", err.Error())
	}
}

func (a *API) recoverRequest(w http.ResponseWriter, r *http.Request) error {
	if recovered := recover(); recovered != nil {
		a.Logger.Error("request panic", "method", r.Method)
		return writeError(w, core.ErrUnavailable)
	}
	return nil
}

func (a *API) route(pattern string, protected bool, action Action) bool {
	a.router.Handle(pattern, endpoint{api: a, action: action, protected: protected})
	return true
}
