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

type HallHandler struct {
	HallService *services.HallService
}

func NewHallHandler(hallService *services.HallService) *HallHandler {
	return &HallHandler{HallService: hallService}
}

func (h *HallHandler) RegisterRoutes(admin *middleware.AdminMiddleware) *http.ServeMux {

	r := http.NewServeMux()

	r.HandleFunc("GET /{id}", middleware.Logging(h.GetHall))
	r.HandleFunc("GET /", middleware.Logging(h.GetHallsPaginated))

	withAdmin := func(next http.HandlerFunc) http.HandlerFunc {
		return middleware.Auth(admin.RequireAdmin(next))
	}

	r.HandleFunc("POST /", middleware.Logging(withAdmin(h.CreateHall)))
	r.HandleFunc("PUT /{id}", middleware.Logging(withAdmin(h.UpdateHall)))
	r.HandleFunc("DELETE /{id}", middleware.Logging(withAdmin(h.DeleteHall)))

	return r
}

func (h *HallHandler) GetHall(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	idString := r.PathValue("id")

	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing hall id. %s", err), http.StatusBadRequest)
		return
	}
	hallID := uint(u64)

	hall, err := h.HallService.GetHall(ctx, hallID)
	if err != nil {
		if errors.Is(err, services.ErrHallNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error getting hall. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, hall)
}

func (h *HallHandler) GetHallsPaginated(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	response, err := h.HallService.GetHallsPaginated(ctx, pageStr, limitStr)
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

func (h *HallHandler) CreateHall(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	var newHall entity.CinemaHall
	err := utils.ParseJSON(r, &newHall)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	createdHall, err := h.HallService.CreateHall(ctx, newHall)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating hall. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, createdHall)
}

func (h *HallHandler) UpdateHall(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	idString := r.PathValue("id")
	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing hall id. %s", err), http.StatusBadRequest)
		return
	}
	hallID := uint(u64)

	var changedHall entity.CinemaHall
	err = utils.ParseJSON(r, &changedHall)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	updatedHall, err := h.HallService.UpdateHall(ctx, hallID, changedHall)
	if err != nil {
		http.Error(w, fmt.Sprintf("error updating hall. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, updatedHall)
}

func (h *HallHandler) DeleteHall(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	idString := r.PathValue("id")
	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing hall id. %s", err), http.StatusBadRequest)
		return
	}
	hallID := uint(u64)

	err = h.HallService.DeleteHall(ctx, hallID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting hall. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "hall deleted")
}
