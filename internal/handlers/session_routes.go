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

type SessionHandler struct {
	SessionService *services.SessionService
}

func NewSessionHandler(sessionService *services.SessionService) *SessionHandler {
	return &SessionHandler{SessionService: sessionService}
}

func (m *SessionHandler) RegisterRoutes(admin *middleware.AdminMiddleware) *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /{id}", middleware.Logging(m.GetSession))
	r.HandleFunc("GET /", middleware.Logging(m.GetSessionsPaginated))

	withAdmin := func(next http.HandlerFunc) http.HandlerFunc {
		return middleware.Auth(admin.RequireAdmin(next))
	}

	r.HandleFunc("POST /", middleware.Logging(withAdmin(m.CreateSession)))
	r.HandleFunc("PUT /{id}", middleware.Logging(withAdmin(m.UpdateSession)))
	r.HandleFunc("DELETE /{id}", middleware.Logging(withAdmin(m.DeleteSession)))

	return r
}

func (m *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	idString := r.PathValue("id")

	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing session id. %s", err), http.StatusBadRequest)
		return
	}
	id := uint(u64)

	session, err := m.SessionService.GetSession(ctx, id)
	if err != nil {
		if errors.Is(err, services.ErrSessionNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error getting session. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, session)
}

func (m *SessionHandler) GetSessionsPaginated(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	response, err := m.SessionService.GetSessionsPaginated(ctx, pageStr, limitStr)
	if err != nil {
		if errors.Is(err, services.ErrNoPageParameter) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (m *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	var newSession entity.Session
	err := utils.ParseJSON(r, &newSession)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	createdSession, err := m.SessionService.CreateSession(ctx, newSession)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating session. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, createdSession)
}

func (m *SessionHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	idString := r.PathValue("id")
	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing session id. %s", err), http.StatusBadRequest)
		return
	}
	id := uint(u64)

	var changedSession entity.Session
	err = utils.ParseJSON(r, &changedSession)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	updatedSession, err := m.SessionService.UpdateSession(ctx, id, changedSession)
	if err != nil {
		http.Error(w, fmt.Sprintf("error updating session. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, updatedSession)
}

func (m *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	idString := r.PathValue("id")
	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing session id. %s", err), http.StatusBadRequest)
		return
	}
	id := uint(u64)

	err = m.SessionService.DeleteSession(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting session. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "session deleted")
}
