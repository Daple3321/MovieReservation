package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Daple3321/MovieReservation/internal/entity"
	"github.com/Daple3321/MovieReservation/internal/middleware"
	"github.com/Daple3321/MovieReservation/internal/services"
	"github.com/Daple3321/MovieReservation/utils"
)

const RequestTimeout time.Duration = time.Second * 3

type UsersHandler struct {
	UserService *services.UserService
}

func NewUsersHandler(userService *services.UserService) *UsersHandler {
	return &UsersHandler{UserService: userService}
}

func (h *UsersHandler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("POST /login", middleware.Logging((h.LoginHandler)))
	r.HandleFunc("POST /register", middleware.Logging((h.RegisterHandler)))

	return r
}

func (h *UsersHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	var u entity.UserDTO
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userId, username, err := h.UserService.Login(ctx, u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// TODO: Get IsAdmin field here...

	token, err := middleware.CreateToken(username, userId)
	if err != nil {
		http.Error(w, errors.New("error creating token").Error(), http.StatusInternalServerError)
	}

	type LoginResponse struct {
		Token    string `json:"token"`
		UserId   uint   `json:"user_id"`
		Username string `json:"username"`
		IsAdmin  bool   `json:"is_admin"`
	}

	resp := LoginResponse{
		Token:    token,
		UserId:   userId,
		Username: username,
		IsAdmin:  false,
	}

	utils.WriteJSONResponse(w, http.StatusOK, resp)
}

func (h *UsersHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	var u entity.UserDTO
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body. %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userId, err := h.UserService.Register(ctx, u)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateUsername) {
			http.Error(w, fmt.Sprintf("username already exists. %s", err), http.StatusConflict)
			return
		} else if errors.Is(err, services.ErrInvalidCredentials) {
			http.Error(w, fmt.Sprintf("error invalid credentials. %s", err), http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("error registering user. %s", err), http.StatusInternalServerError)
		return
	}

	resp := struct {
		UserId uint `json:"user_id"`
	}{
		UserId: userId,
	}

	utils.WriteJSONResponse(w, http.StatusCreated, resp)
}
