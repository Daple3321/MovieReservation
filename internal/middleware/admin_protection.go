package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Daple3321/MovieReservation/internal/services"
)

type AdminMiddleware struct {
	userService *services.UserService
}

func NewAdminMiddleware(userService *services.UserService) *AdminMiddleware {
	return &AdminMiddleware{userService: userService}
}

func (a *AdminMiddleware) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
		defer cancel()

		userId, err := GetUserIdFromCtx(ctx)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := a.userService.GetByID(ctx, userId)
		if err != nil {
			if errors.Is(err, services.ErrUserNotFound) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if user.Role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
