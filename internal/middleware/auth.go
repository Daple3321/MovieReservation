package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type ContextKey string

const ContextUserIdKey ContextKey = "userId"

func GetUserIdFromCtx(ctx context.Context) (uint, error) {
	userIdVal := ctx.Value(ContextUserIdKey)
	if userIdVal == nil {
		return 0, errors.New("user id not in context")
	}
	userId, ok := userIdVal.(uint)
	if !ok {
		return 0, errors.New("user id invalid type")
	}

	return userId, nil
}

func CreateToken(username string, userId uint) (string, error) {
	// sets secretKey for the first time if it was nil
	if len(secretKey) == 0 {
		secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
		//fmt.Printf("secret key == nil. Getting from env: %s\n", secretKey)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"name":   username,
			"userId": userId,
			"exp":    time.Now().Add(time.Hour * 3).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (uint, error) {
	if len(secretKey) == 0 {
		secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
		//fmt.Printf("secret key == nil. Getting from env: %s\n", secretKey)
	}

	token, err := jwt.Parse(tokenString,
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
	if err != nil {
		return 0, err
	}
	if token.Method != jwt.SigningMethodHS256 {
		return 0, errors.New("wrong JWT token signing method")
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {

		if userID, ok := claims["userId"].(float64); ok {
			return uint(userID), nil
		}

		return 0, fmt.Errorf("user_id not found or is of invalid type in claims")
	}

	return 0, fmt.Errorf("invalid token claims")
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "token not provided.", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		userId, err := VerifyToken(tokenString)
		if err != nil {
			slog.Info("error creating token", "userId", userId, "err", err)
			http.Error(w, errors.New("unauthorized").Error(), http.StatusUnauthorized)
			return
		}

		// pass new context with added userId key
		ctx := context.WithValue(r.Context(), ContextUserIdKey, userId)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
