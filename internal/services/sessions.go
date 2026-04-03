package services

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Daple3321/MovieReservation/internal/entity"
	"gorm.io/gorm"
)

var ErrSessionNotFound = errors.New("session not found")
var ErrHallNotFound = errors.New("hall not found")
var ErrMovieNotFound = errors.New("movie not found")

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) GetSession(ctx context.Context, sessionID uint) (*entity.Session, error) {
	var session entity.Session

	// BUG: This whole thing is given to a user.
	// User can see all tickets and seats.
	// get session with movie, cinema hall, seats, and tickets
	result := s.db.
		WithContext(ctx).
		Preload("Movie").
		Preload("CinemaHall").
		Preload("Seats").
		Preload("Tickets").
		First(&session, sessionID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, result.Error
	}

	return &session, nil
}

func (s *SessionService) GetSessionsPaginated(ctx context.Context, pageStr string, limitStr string) (*entity.PaginatedResponse, error) {
	pageStr = strings.TrimSpace(pageStr)
	if pageStr == "" {
		return nil, ErrNoPageParameter
	}
	limitStr = strings.TrimSpace(limitStr)

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // Default to page 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10 // Default to 10 items per page
	}

	// BUG: This whole thing is given to a user.
	// User can see all tickets and seats.
	// get sessions
	sessions := []entity.Session{}
	result := s.db.
		WithContext(ctx).
		Limit(limit).
		Offset((page - 1) * limit).
		Preload("Movie").
		Preload("CinemaHall").
		Preload("Seats").
		Preload("Tickets").
		Find(&sessions)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, result.Error
	}

	var totalItems int64
	s.db.WithContext(ctx).Model(&entity.Session{}).Count(&totalItems)
	totalPages := (int(totalItems) + limit - 1) / limit

	response := entity.PaginatedResponse{
		Items:      sessions,
		Page:       page,
		Limit:      limit,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
	}

	return &response, nil
}

func (s *SessionService) CreateSession(ctx context.Context, session entity.Session) (*entity.Session, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("session create: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	// validate hall exists
	hall := entity.CinemaHall{}
	result := tx.First(&hall, session.HallID)
	if result.Error != nil {
		slog.Error("session create: hall not found", "err", result.Error)
		return nil, ErrHallNotFound
	}

	// validate movie exists
	movie := entity.Movie{}
	result = tx.First(&movie, session.MovieID)
	if result.Error != nil {
		slog.Error("session create: movie not found", "err", result.Error)
		return nil, ErrMovieNotFound
	}

	// create session
	if err := tx.Create(&session).Error; err != nil {
		slog.Error("session create: insert", "err", err)
		return nil, ErrInternalServer
	}

	// create seats
	seats := []entity.Seat{}
	for i := 0; i < hall.Width; i++ {
		for j := 0; j < hall.Height; j++ {
			seat := entity.Seat{
				SessionID: session.ID,
				X:         i,
				Y:         j,
			}
			seats = append(seats, seat)
		}
	}

	if err := tx.Create(&seats).Error; err != nil {
		slog.Error("session create: insert seats", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("session create: commit", "err", err)
		return nil, ErrInternalServer
	}

	return &session, nil
}

func (s *SessionService) UpdateSession(ctx context.Context, sessionID uint, session entity.Session) (*entity.Session, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("session update: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	// validate session exists
	sessionToUpdate := entity.Session{}
	result := tx.First(&sessionToUpdate, sessionID)
	if result.Error != nil {
		slog.Error("session update: session not found", "err", result.Error)
		return nil, ErrSessionNotFound
	}

	// validate hall exists
	hall := entity.CinemaHall{}
	result = tx.First(&hall, sessionToUpdate.HallID)
	if result.Error != nil {
		slog.Error("session update: hall not found", "err", result.Error)
		return nil, ErrHallNotFound
	}

	// validate movie exists
	movie := entity.Movie{}
	result = tx.First(&movie, sessionToUpdate.MovieID)
	if result.Error != nil {
		slog.Error("session update: movie not found", "err", result.Error)
		return nil, ErrMovieNotFound
	}

	// update session
	if err := tx.Save(&sessionToUpdate).Error; err != nil {
		slog.Error("session update: insert", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("session update: commit", "err", err)
		return nil, ErrInternalServer
	}

	return &sessionToUpdate, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, sessionID uint) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("session delete: begin tx", "err", tx.Error)
		return ErrInternalServer
	}
	defer tx.Rollback()

	if err := tx.Delete(&entity.Session{}, sessionID).Error; err != nil {
		slog.Error("session delete: insert", "err", err)
		return ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("session delete: commit", "err", err)
		return ErrInternalServer
	}

	return nil
}
