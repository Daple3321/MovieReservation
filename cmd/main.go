package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/Daple3321/MovieReservation/internal/entity"
	"github.com/Daple3321/MovieReservation/internal/handlers"
	"github.com/Daple3321/MovieReservation/internal/middleware"
	"github.com/Daple3321/MovieReservation/internal/services"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		slog.Warn("no .env file found, using process environment", "envPath", envPath)
	}

	logFile, err := SetupLogger()
	if err != nil {
		log.Fatal("logger setup error", err)
	}
	defer logFile.Close()

	if err := ValidateEnvVars(); err != nil {
		slog.Error("error validating env vars", "err", err)
		os.Exit(1)
	}

	db, err := SetupDB()
	if err != nil {
		slog.Error("error setting up db", "err", err)
		os.Exit(1)
	}

	err = db.AutoMigrate(
		&entity.Movie{},
		&entity.User{},
		&entity.CinemaHall{},
		&entity.Session{},
		&entity.Ticket{},
	)
	if err != nil {
		slog.Error("error migrating db", "err", err)
		os.Exit(1)
	}

	userService := services.NewUserService(db)
	usersHandler := handlers.NewUsersHandler(userService)
	authRouter := usersHandler.RegisterRoutes()

	adminMiddleware := middleware.NewAdminMiddleware(userService)

	router := http.NewServeMux()

	router.Handle("GET /auth/", http.StripPrefix("/auth", authRouter))

	// Admin-only example: JWT must be valid (Auth) before RequireAdmin loads the user from DB.
	router.HandleFunc("POST /movies", middleware.Logging(
		middleware.Auth(adminMiddleware.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "movie create not implemented", http.StatusNotImplemented)
		})),
	))

	serverIP := getEnv("SERVERIP", "0.0.0.0")
	serverPort := os.Getenv("SERVERPORT")
	slog.Info("Listening on:", "ip", serverIP, "port", serverPort)
	if err := http.ListenAndServe(serverIP+":"+serverPort, router); err != nil {
		slog.Error("error starting http server", "err", err)
		os.Exit(1)
	}
}

func SetupDB() (*gorm.DB, error) {

	dbUser := os.Getenv("MOVIEDB_USERNAME")
	dbPassword := os.Getenv("MOVIEDB_PASSWORD")

	dsn := fmt.Sprintf("host=localhost user=%s dbname=MovieTheater password=%s sslmode=disable", dbUser, dbPassword)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ValidateEnvVars() error {
	vars := []string{
		"SERVERIP",
		"SERVERPORT",
		"MOVIEDB_USERNAME",
		"MOVIEDB_PASSWORD",
		"JWT_SECRET_KEY",
		"LOG_LEVEL",
	}

	for _, v := range vars {
		if os.Getenv(v) == "" {
			return fmt.Errorf("env var %s not set", v)
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func SetupLogger() (*os.File, error) {
	workDir, _ := os.Getwd()
	logPath := path.Join(workDir, "server.log")
	//os.WriteFile(logPath, []byte{}, os.ModeAppend)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("failed to open log file", "err", err)
		return nil, err
	}

	logLevelString := getEnv("LOG_LEVEL", "debug")
	logLevel := -4
	switch logLevelString {
	case "debug":
		logLevel = -4
	case "info":
		logLevel = 0
	case "warn":
		logLevel = 4
	case "error":
		logLevel = 8
	}

	opts := slog.HandlerOptions{
		Level: slog.Level(logLevel),
	}
	w := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewTextHandler(w, &opts))
	slog.SetDefault(logger)

	return logFile, nil
}
