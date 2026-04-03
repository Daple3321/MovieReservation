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
var ErrSessionHallImmutable = errors.New("session hall cannot be changed after creation; delete and recreate the session")
var ErrSessionMissingRefs = errors.New("hall_id and movie_id are required")

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) GetSession(ctx context.Context, sessionID uint) (*entity.Session, error) {
	var session entity.Session

	result := s.db.
		WithContext(ctx).
		Preload("Movie").
		Preload("CinemaHall").
		Preload("Seats").
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

	sessions := []entity.Session{}
	result := s.db.
		WithContext(ctx).
		Limit(limit).
		Offset((page - 1) * limit).
		Preload("Movie").
		Preload("CinemaHall").
		Preload("Seats").
		Find(&sessions)

	if result.Error != nil {
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
	if session.HallID == 0 || session.MovieID == 0 {
		return nil, ErrSessionMissingRefs
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("session create: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	hall := entity.CinemaHall{}
	if err := tx.First(&hall, session.HallID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHallNotFound
		}
		slog.Error("session create: load hall", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.First(&entity.Movie{}, session.MovieID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMovieNotFound
		}
		slog.Error("session create: load movie", "err", err)
		return nil, ErrInternalServer
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

	var out entity.Session
	if err := s.db.WithContext(ctx).
		Preload("Movie").
		Preload("CinemaHall").
		Preload("Seats").
		First(&out, session.ID).Error; err != nil {
		return nil, err
	}

	return &out, nil
}

func (s *SessionService) UpdateSession(ctx context.Context, sessionID uint, changed entity.Session) (*entity.Session, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("session update: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	sessionToUpdate := entity.Session{}
	if err := tx.First(&sessionToUpdate, sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		slog.Error("session update: load session", "err", err)
		return nil, ErrInternalServer
	}

	if changed.HallID != 0 && changed.HallID != sessionToUpdate.HallID {
		return nil, ErrSessionHallImmutable
	}

	if changed.MovieID != 0 && changed.MovieID != sessionToUpdate.MovieID {
		if err := tx.First(&entity.Movie{}, changed.MovieID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrMovieNotFound
			}
			slog.Error("session update: validate movie", "err", err)
			return nil, ErrInternalServer
		}
		sessionToUpdate.MovieID = changed.MovieID
	}

	sessionToUpdate.Price = changed.Price
	sessionToUpdate.StartTime = changed.StartTime

	if err := tx.Save(&sessionToUpdate).Error; err != nil {
		slog.Error("session update: save", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("session update: commit", "err", err)
		return nil, ErrInternalServer
	}

	var out entity.Session
	if err := s.db.WithContext(ctx).
		Preload("Movie").
		Preload("CinemaHall").
		Preload("Seats").
		First(&out, sessionID).Error; err != nil {
		return nil, err
	}

	return &out, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, sessionID uint) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("session delete: begin tx", "err", tx.Error)
		return ErrInternalServer
	}
	defer tx.Rollback()

	if err := tx.First(&entity.Session{}, sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		slog.Error("session delete: load session", "err", err)
		return ErrInternalServer
	}

	if err := tx.Where("session_id = ?", sessionID).Delete(&entity.Ticket{}).Error; err != nil {
		slog.Error("session delete: tickets", "err", err)
		return ErrInternalServer
	}
	if err := tx.Where("session_id = ?", sessionID).Delete(&entity.Seat{}).Error; err != nil {
		slog.Error("session delete: seats", "err", err)
		return ErrInternalServer
	}
	if err := tx.Delete(&entity.Session{}, sessionID).Error; err != nil {
		slog.Error("session delete: session", "err", err)
		return ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("session delete: commit", "err", err)
		return ErrInternalServer
	}

	return nil
}
