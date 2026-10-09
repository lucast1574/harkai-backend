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
		{"POST /v1/analysis/text", true, a.analyzeText},
		{"POST /v1/analysis/audio", true, a.analyzeAudio},
		{"GET /v1/media-tools/source", false, a.mediaSource},
		{"GET /openapi.json", false, a.specification},
		{"POST /v1/auth/register", false, a.register},
		{"POST /v1/auth/login", false, a.login},
		{"POST /v1/auth/google", false, a.googleLogin},
		{"POST /v1/auth/refresh", false, a.refresh},
		{"POST /v1/auth/logout", true, a.logout},
		{"PUT /v1/me/devices", true, a.device},
		{"DELETE /v1/me/devices/{deviceID}", true, a.deleteDevice},
		{"GET /v1/me", true, a.me},
		{"PATCH /v1/me", true, a.profile},
		{"PUT /v1/me/preferences", true, a.preferences},
		{"GET /v1/me/incidents", true, a.myIncidents},
		{"GET /v1/archive/districts", false, a.districts},
		{"GET /v1/archive/incidents", false, a.archive},
		{"GET /v1/incidents/{id}/comments", false, a.comments},
		{"POST /v1/incidents/{id}/comments", true, a.comment},
		{"DELETE /v1/incidents/{id}/comments/{commentID}", true, a.deleteComment},
		{"GET /v1/incidents", false, a.listIncidents},
		{"POST /v1/incidents", true, a.report},
		{"GET /v1/incidents/{id}", false, a.incident},
		{"PATCH /v1/incidents/{id}/status", true, a.incidentStatus},
		{"GET /v1/alerts", true, a.alerts},
		{"GET /v1/admin/incidents", true, a.adminIncidents},
		{"GET /v1/gov/incidents", true, a.govIncidents},
		{"POST /v1/incidents/{id}/confirm", true, a.confirmReport},
		{"GET /v1/admin/users", true, a.adminUsers},
		{"PATCH /v1/admin/users/{id}/role", true, a.assignRole},
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
	return Result{Body: map[string]any{"api_version": "v1", "incident_types": core.IncidentTypes(), "categories": core.Categories(), "roles": core.Roles(), "verification": "one_other_user", "capabilities": map[string]bool{"email_auth": true, "google_auth": a.Service.VerifyGoogle != nil, "geo_reports": true, "media": a.Service.ModerateImage != nil, "text_rules": true, "audio": a.Service.TranscribeAudio != nil, "pull_alerts": true, "push_notifications": a.Service.PushEnabled, "payments": false, "generative_ai": false}}}, nil
}
