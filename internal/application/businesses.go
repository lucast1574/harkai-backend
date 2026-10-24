package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

type BusinessStore interface {
	CreateBusiness(context.Context, core.BusinessPoint) (core.BusinessPoint, error)
	UpdateBusiness(context.Context, core.BusinessPoint) (core.BusinessPoint, error)
	BusinessByID(context.Context, string) (core.BusinessPoint, error)
	ListBusinesses(context.Context, core.IncidentQuery, string, bool, time.Time) (core.BusinessPage, error)
	ModerateBusiness(context.Context, string, string, time.Time) (core.BusinessPoint, error)
	SetBusinessEnabled(context.Context, string, string, bool, time.Time) (core.BusinessPoint, error)
	ApplyBusinessEntitlement(context.Context, core.BusinessEntitlement, time.Time) error
	QueueBusinessNotice(context.Context, core.BusinessNotice) error
	BusinessNotices(context.Context, string, string, string, time.Time) ([]core.BusinessNotice, error)
	AckBusinessNotice(context.Context, string, string, string) error
}
type BusinessBillingVerifier interface {
	VerifyBusinessEntitlement(context.Context, []byte, string) (core.BusinessEntitlement, error)
}

func (s Service) SaveBusiness(ctx context.Context, u core.User, id string, in core.BusinessInput) (core.BusinessPoint, error) {
	if s.Businesses == nil {
		return core.BusinessPoint{}, core.ErrUnavailable
	}
	now := s.Now()
	next, err := core.BuildBusiness(in, s.ID(), u.ID, now)
	if err != nil {
		return next, err
	}
	if err = s.ValidateMedia(ctx, u.ID, in.MediaID); err != nil {
		return next, err
	}
	if err = s.Limits.Allow(ctx, "business-save:"+u.ID, 20, time.Hour, now); err != nil {
		return next, err
	}
	if id == "" {
		return s.Businesses.CreateBusiness(ctx, next)
	}
	if !core.ValidID(id) {
		return next, core.ErrInvalid
	}
	old, err := s.Businesses.BusinessByID(ctx, id)
	if err != nil {
		return next, err
	}
	if old.OwnerID != u.ID {
		return next, core.ErrForbidden
	}
	next.ID, next.CreatedAt = old.ID, old.CreatedAt
	// Editing public content requires moderation again, but cannot mint billing rights.
	next.BillingStatus, next.PaidUntil, next.SubscriptionID = old.BillingStatus, old.PaidUntil, old.SubscriptionID
	next.BillingEventAt, next.BillingEventID = old.BillingEventAt, old.BillingEventID
	return s.Businesses.UpdateBusiness(ctx, next)
}
func (s Service) Business(ctx context.Context, u core.User, id string) (core.BusinessPoint, error) {
	if s.Businesses == nil {
		return core.BusinessPoint{}, core.ErrUnavailable
	}
	if !core.ValidID(id) {
		return core.BusinessPoint{}, core.ErrInvalid
	}
	point, err := s.Businesses.BusinessByID(ctx, id)
	if err != nil {
		return point, err
	}
	if point.OwnerID != u.ID && u.Role != "admin" && !point.Public(s.Now()) {
		return core.BusinessPoint{}, core.ErrNotFound
	}
	return core.PublicBusiness(point), nil
}
func (s Service) FindBusinesses(ctx context.Context, u core.User, q core.IncidentQuery, own, admin bool) (core.BusinessPage, error) {
	if s.Businesses == nil {
		return core.BusinessPage{}, core.ErrUnavailable
	}
	if admin && u.Role != "admin" {
		return core.BusinessPage{}, core.ErrForbidden
	}
	validated, err := core.ValidateQuery(q)
	if err != nil {
		return core.BusinessPage{}, err
	}
	owner := ""
	if own {
		owner = u.ID
	}
	page, err := s.Businesses.ListBusinesses(ctx, validated, owner, admin, s.Now())
	for n := range page.Items {
		page.Items[n] = core.PublicBusiness(page.Items[n])
	}
	return page, err
}
func (s Service) ReviewBusiness(ctx context.Context, u core.User, id, status string) (core.BusinessPoint, error) {
	if s.Businesses == nil {
		return core.BusinessPoint{}, core.ErrUnavailable
	}
	if u.Role != "admin" {
		return core.BusinessPoint{}, core.ErrForbidden
	}
	if !core.ValidID(id) || (status != "approved" && status != "rejected") {
		return core.BusinessPoint{}, core.ErrInvalid
	}
	return s.Businesses.ModerateBusiness(ctx, id, status, s.Now())
}
func (s Service) EnableBusiness(ctx context.Context, u core.User, id string, enabled bool) (core.BusinessPoint, error) {
	if s.Businesses == nil {
		return core.BusinessPoint{}, core.ErrUnavailable
	}
	if !core.ValidID(id) {
		return core.BusinessPoint{}, core.ErrInvalid
	}
	return s.Businesses.SetBusinessEnabled(ctx, id, u.ID, enabled, s.Now())
}
func (s Service) BusinessBillingEvent(ctx context.Context, body []byte, signature string) error {
	if s.Billing == nil || s.Businesses == nil {
		return core.ErrUnavailable
	}
	e, err := s.Billing.VerifyBusinessEntitlement(ctx, body, signature)
	if err != nil {
		return err
	}
	if err = core.ValidateBusinessEntitlement(e, s.Now()); err != nil {
		return err
	}
	return s.Businesses.ApplyBusinessEntitlement(ctx, e, s.Now())
}
