package httpapi

import (
	"net/http"
	"strconv"

	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
)

type ticketRequester struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type ticketView struct {
	core.SupportTicket
	Screenshots []string         `json:"images"`
	Requester   *ticketRequester `json:"requester,omitempty"`
}

func viewTicket(ticket core.SupportTicket, admin bool) ticketView {
	view := ticketView{SupportTicket: ticket, Screenshots: []string{}}
	for i := range ticket.Images {
		view.Screenshots = append(view.Screenshots, "/api/backend/support/tickets/"+ticket.ID+"/images/"+strconv.Itoa(i))
	}
	if admin {
		view.Requester = &ticketRequester{Name: ticket.RequesterName, Email: ticket.RequesterEmail}
	}
	return view
}

func (a *API) ticketCreate(r *http.Request, user core.User) (Result, error) {
	var input struct {
		RequestID   string   `json:"requestId"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Images      []string `json:"images"`
	}
	if err := decodeSized(r, &input, 5<<20); err != nil {
		return Result{}, err
	}
	ticket, err := a.Service.CreateSupportTicket(r.Context(), user, input.RequestID, input.Title, input.Description, func() ([]core.SupportImage, error) { return supportImages(input.Images) })
	return Result{Status: 201, Body: viewTicket(ticket, false)}, err
}

func (a *API) ticketList(r *http.Request, user core.User) (Result, error) {
	return a.listTickets(r, user, false)
}
func (a *API) ticketAdminList(r *http.Request, user core.User) (Result, error) {
	return a.listTickets(r, user, true)
}
func (a *API) listTickets(r *http.Request, user core.User, admin bool) (Result, error) {
	page := int64(1)
	if raw := r.URL.Query().Get("page"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return Result{}, core.ErrInvalid
		}
		page = value
	}
	items, total, err := a.Service.SupportTickets(r.Context(), user, admin, application.TicketQuery{State: r.URL.Query().Get("state"), Page: page})
	if err != nil {
		return Result{}, err
	}
	views := make([]ticketView, 0, len(items))
	for _, ticket := range items {
		views = append(views, viewTicket(ticket, admin))
	}
	return Result{Body: map[string]any{"items": views, "total": total, "page": page, "pages": (total + 19) / 20}}, nil
}

func (a *API) ticketUpdate(r *http.Request, user core.User) (Result, error) {
	var input struct {
		State   string `json:"state"`
		Note    string `json:"note"`
		Version *int64 `json:"version"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	if input.Version == nil {
		return Result{}, core.ErrInvalid
	}
	ticket, err := a.Service.UpdateSupportTicket(r.Context(), user, r.PathValue("id"), input.State, input.Note, *input.Version)
	return Result{Body: viewTicket(ticket, true)}, err
}

func (a *API) ticketImage(r *http.Request, user core.User) (Result, error) {
	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		return Result{}, core.ErrNotFound
	}
	image, err := a.Service.SupportImage(r.Context(), user, r.PathValue("id"), index)
	return Result{Bytes: image.Data, Headers: http.Header{"Content-Type": {image.MIME}, "Cache-Control": {"private, no-store"}, "Content-Security-Policy": {"default-src 'none'; sandbox"}}}, err
}
