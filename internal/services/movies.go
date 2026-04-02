package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Daple3321/MovieReservation/internal/entity"
	"gorm.io/gorm"
)

type MovieService struct {
	db *gorm.DB
}

func NewMovieService(db *gorm.DB) *MovieService {
	return &MovieService{db: db}
}

func (m *MovieService) GetMovie(ctx context.Context, movieID uint) (*entity.Movie, error) {
	var movie entity.Movie

	result := m.db.
		WithContext(ctx).
		Preload("Sessions").
		Preload("Users").
		First(&movie, movieID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}

	return &movie, nil
}

func (m *MovieService) CreateMovie(ctx context.Context, movie entity.Movie) (*entity.Movie, error) {
	tx := m.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("movie create: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	if err := tx.Create(&movie).Error; err != nil {
		slog.Error("movie create: insert", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("movie create: commit", "err", err)
		return nil, ErrInternalServer
	}

	return &movie, nil
}

func (m *MovieService) UpdateMovie(ctx context.Context, changedMovie entity.Movie) (*entity.Movie, error) {
	tx := m.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("movie update: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	if err := tx.Save(&changedMovie).Error; err != nil {
		slog.Error("movie update: insert", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("movie update: commit", "err", err)
		return nil, ErrInternalServer
	}

	return &changedMovie, nil
}

func (m *MovieService) DeleteMovie(ctx context.Context, movieId uint) error {
	tx := m.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("movie delete: begin tx", "err", tx.Error)
		return ErrInternalServer
	}
	defer tx.Rollback()

	if err := tx.Delete(&entity.Movie{}, movieId).Error; err != nil {
		slog.Error("movie delete: insert", "err", err)
		return ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("movie delete: commit", "err", err)
		return ErrInternalServer
	}

	return nil
}
