package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

type Users interface {
	CreateUser(context.Context, core.User) (core.User, error)
	UserByEmail(context.Context, string) (core.User, error)
	UserByID(context.Context, string) (core.User, error)
	UpdatePreferences(context.Context, string, core.Preferences) (core.User, error)
	UpdateName(context.Context, string, string) (core.User, error)
}

type Sessions interface {
	CreateSession(context.Context, core.Session) error
	SessionByAccess(context.Context, string, time.Time) (core.Session, error)
	RotateSession(context.Context, string, core.Session, time.Time) (core.Session, error)
	DeleteSession(context.Context, string) error
}

type Incidents interface {
	CreateIncident(context.Context, core.Incident) (core.Incident, error)
	IncidentByID(context.Context, string) (core.Incident, error)
	ListIncidents(context.Context, core.IncidentQuery, time.Time) (core.IncidentPage, error)
	SetIncidentStatus(context.Context, string, string, string, time.Time) (core.Incident, error)
	ClaimCooldown(context.Context, string, string, time.Time) error
}

type Limits interface {
	Allow(context.Context, string, int, time.Duration, time.Time) error
}

type Service struct {
	Users             Users
	Sessions          Sessions
	Incidents         Incidents
	Limits            Limits
	Media             MediaStore
	Photos            PhotoStore
	Directory         ContactsStore
	NormalizeImage    func([]byte) ([]byte, error)
	Now               func() time.Time
	ID                func() string
	Token             func() (string, error)
	HashPassword      func(string) (string, error)
	CheckPassword     func(string, string) bool
	HashToken         func(string) string
	DummyPasswordHash string
}
