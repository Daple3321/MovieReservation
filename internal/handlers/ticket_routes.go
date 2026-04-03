package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Daple3321/MovieReservation/internal/entity"
	"github.com/Daple3321/MovieReservation/internal/middleware"
	"github.com/Daple3321/MovieReservation/internal/services"
	"github.com/Daple3321/MovieReservation/utils"
)

type TicketHandler struct {
	TicketService *services.TicketService
}

func NewTicketHandler(ticketService *services.TicketService) *TicketHandler {
	return &TicketHandler{TicketService: ticketService}
}

func (t *TicketHandler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /{id}", middleware.Logging(middleware.Auth(t.GetTicket)))
	r.HandleFunc("GET /", middleware.Logging(middleware.Auth(t.GetTicketsPaginated)))

	r.HandleFunc("POST /", middleware.Logging(middleware.Auth((t.BuyTicket))))
	r.HandleFunc("DELETE /{id}", middleware.Logging(middleware.Auth((t.CancelTicket))))

	return r
}

func (t *TicketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	userId, err := middleware.GetUserIdFromCtx(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idString := r.PathValue("id")

	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing ticket id. %s", err), http.StatusBadRequest)
		return
	}
	id := uint(u64)

	ticket, err := t.TicketService.GetTicket(ctx, userId, id)
	if err != nil {
		if errors.Is(err, services.ErrTicketNotFound) {
			http.Error(w, "ticket not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error getting ticket. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, ticket)
}

func (t *TicketHandler) GetTicketsPaginated(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	userId, err := middleware.GetUserIdFromCtx(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	response, err := t.TicketService.GetTicketsPaginated(ctx, userId, pageStr, limitStr)
	if err != nil {
		if errors.Is(err, services.ErrNoPageParameter) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (t *TicketHandler) BuyTicket(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	userId, err := middleware.GetUserIdFromCtx(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var newTicket entity.Ticket
	err = utils.ParseJSON(r, &newTicket)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	createdTicket, err := t.TicketService.BuyTicket(ctx, userId, newTicket)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidTicketRequest),
			errors.Is(err, services.ErrSeatSessionMismatch):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, services.ErrSessionMismatch),
			errors.Is(err, services.ErrSeatNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, services.ErrSeatNotAvailable):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, fmt.Sprintf("error buying ticket. %s", err), http.StatusInternalServerError)
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, createdTicket)
}

func (t *TicketHandler) CancelTicket(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	userId, err := middleware.GetUserIdFromCtx(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idString := r.PathValue("id")

	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing ticket id. %s", err), http.StatusBadRequest)
		return
	}
	id := uint(u64)

	err = t.TicketService.CancelTicket(ctx, userId, id)
	if err != nil {
		if errors.Is(err, services.ErrTicketNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, services.ErrSeatNotFound) || errors.Is(err, services.ErrSeatNotAvailable) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("error canceling ticket. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "ticket canceled")
}
