package middleware

import (
	"log/slog"
	"net/http"
)

func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attrs := slog.Group("Request", "method", r.Method, "requestURI", r.RequestURI, "remoteAddr", r.RemoteAddr)
		slog.Info("Request recieved", attrs)
		next(w, r)
	}
}
