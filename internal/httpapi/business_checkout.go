package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

func (a *API) businessCheckout(r *http.Request, u core.User) (Result, error) {
	p, err := a.Service.CheckoutBusiness(r.Context(), u, r.PathValue("id"))
	return Result{Body: p}, err
}
