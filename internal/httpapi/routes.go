package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

func (a *API) routes() int {
	routes := []struct {
		pattern   string
		protected bool
		action    Action
	}{
		{"GET /health/live", false, a.live},
		{"GET /health/ready", false, a.ready},
		{"GET /v1/meta", false, a.meta},
		{"GET /openapi.json", false, a.specification},
		{"POST /v1/auth/register", false, a.register},
		{"POST /v1/auth/login", false, a.login},
		{"POST /v1/auth/refresh", false, a.refresh},
		{"POST /v1/auth/logout", true, a.logout},
		{"GET /v1/me", true, a.me},
		{"PATCH /v1/me", true, a.profile},
		{"PUT /v1/me/preferences", true, a.preferences},
		{"GET /v1/me/incidents", true, a.myIncidents},
		{"GET /v1/incidents", false, a.listIncidents},
		{"POST /v1/incidents", true, a.report},
		{"GET /v1/incidents/{id}", false, a.incident},
		{"PATCH /v1/incidents/{id}/status", true, a.incidentStatus},
		{"GET /v1/alerts", true, a.alerts},
		{"GET /v1/admin/incidents", true, a.adminIncidents},
		{"POST /v1/media", true, a.upload},
		{"GET /v1/media/{id}", false, a.download},
		{"GET /v1/emergency-contacts", false, a.contacts},
		{"PUT /v1/admin/emergency-contacts", true, a.saveContacts},
	}
	for _, r := range routes {
		_ = a.route(r.pattern, r.protected, r.action)
	}
	return len(routes)
}

func (a *API) live(_ *http.Request, _ core.User) (Result, error) {
	return Result{Body: map[string]string{"status": "ok", "service": "harkai-backend"}}, nil
}

func (a *API) ready(r *http.Request, _ core.User) (Result, error) {
	if err := a.Ping(r.Context()); err != nil {
		return Result{}, core.ErrUnavailable
	}
	return Result{Body: map[string]string{"status": "ready"}}, nil
}

func (a *API) meta(_ *http.Request, _ core.User) (Result, error) {
	return Result{Body: map[string]any{"api_version": "v1", "incident_types": core.IncidentTypes(), "capabilities": map[string]bool{"email_auth": true, "geo_reports": true, "media": true, "pull_alerts": true, "push_notifications": false, "payments": false, "ai": false}}}, nil
}
