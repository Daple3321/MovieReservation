package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Daple3321/MovieReservation/internal/entity"
	"github.com/Daple3321/MovieReservation/internal/middleware"
	"github.com/Daple3321/MovieReservation/internal/services"
	"github.com/Daple3321/MovieReservation/utils"
)

type MovieHandler struct {
	MovieService *services.MovieService
}

func NewMovieHandler(movieService *services.MovieService) *MovieHandler {
	return &MovieHandler{MovieService: movieService}
}

func (m *MovieHandler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /{id}", middleware.Logging((m.GetMovie)))
	r.HandleFunc("POST /", middleware.Logging((m.CreateMovie)))
	r.HandleFunc("PUT /", middleware.Logging((m.UpdateMovie)))
	r.HandleFunc("DELETE /{id}", middleware.Logging((m.DeleteMovie)))

	return r
}

func (m *MovieHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	idString := r.PathValue("id")

	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing movie id. %s", err), http.StatusBadRequest)
		return
	}
	id := uint(u64)

	movie, err := m.MovieService.GetMovie(ctx, id)

	utils.WriteJSONResponse(w, http.StatusOK, movie)
}

func (m *MovieHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	var newMovie entity.Movie
	err := utils.ParseJSON(r, &newMovie)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	newMovie.ReleaseDate = time.Now()

	createdMovie, err := m.MovieService.CreateMovie(ctx, newMovie)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating movie. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, createdMovie)
}

func (m *MovieHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	// idString := r.PathValue("id")

	// u64, err := strconv.ParseUint(idString, 10, 64)
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("error parsing movie id. %s", err), http.StatusBadRequest)
	// 	return
	// }
	// id := uint(u64)

	var changedMovie entity.Movie
	err := utils.ParseJSON(r, &changedMovie)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	updatedMovie, err := m.MovieService.UpdateMovie(ctx, changedMovie)
	if err != nil {
		http.Error(w, fmt.Sprintf("error updating movie. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, updatedMovie)
}

func (m *MovieHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	idString := r.PathValue("id")

	u64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing movie id. %s", err), http.StatusBadRequest)
		return
	}
	id := uint(u64)

	err = m.MovieService.DeleteMovie(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting movie. %s", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "movie deleted")
}
