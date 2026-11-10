package application

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

type BusinessCheckoutStore interface {
	BusinessCheckoutByID(context.Context, string) (core.BusinessCheckout, error)
	BusinessCheckout(context.Context, core.BusinessCheckout) (core.BusinessCheckout, error)
	RotateBusinessCheckout(context.Context, core.BusinessCheckout, string) error
	SaveBusinessCheckout(context.Context, core.BusinessCheckout, string) error
}
type BusinessCheckoutProvider interface {
	Checkout(context.Context, core.BusinessPoint, core.BusinessCheckout) (core.CheckoutSession, error)
}

func (s Service) CheckoutBusiness(ctx context.Context, u core.User, id string) (core.CheckoutSession, error) {
	if s.Checkout == nil || s.Checkouts == nil || s.Businesses == nil {
		return core.CheckoutSession{}, core.ErrUnavailable
	}
	if !core.ValidID(id) {
		return core.CheckoutSession{}, core.ErrInvalid
	}
	p, err := s.Businesses.BusinessByID(ctx, id)
	if err != nil {
		return core.CheckoutSession{}, err
	}
	if p.OwnerID != u.ID {
		return core.CheckoutSession{}, core.ErrForbidden
	}
	// Approval first: a rejected listing must never be charged.
	if p.Moderation != "approved" || p.BillingStatus == "active" || p.BillingStatus == "refunded" {
		return core.CheckoutSession{}, core.ErrConflict
	}
	if err = s.Limits.Allow(ctx, "business-checkout:"+u.ID, 20, time.Hour, s.Now()); err != nil {
		return core.CheckoutSession{}, err
	}
	for attempt := 0; attempt < 2; attempt++ {
		record, err := s.Checkouts.BusinessCheckout(ctx, core.BusinessCheckout{PointID: id, OwnerID: u.ID, Key: s.ID(), CreatedAt: s.Now()})
		if err != nil {
			return core.CheckoutSession{}, err
		}
		checkout, err := s.Checkout.Checkout(ctx, p, record)
		if err == nil && checkout.ID != "" && record.SessionID == "" {
			err = s.Checkouts.SaveBusinessCheckout(ctx, record, checkout.ID)
		}
		if err != nil || checkout.Status != "expired" {
			return checkout, err
		}
		if err = s.Checkouts.RotateBusinessCheckout(ctx, record, s.ID()); err != nil && !errors.Is(err, core.ErrConflict) {
			return core.CheckoutSession{}, err
		}
	}
	return core.CheckoutSession{}, core.ErrConflict
}
