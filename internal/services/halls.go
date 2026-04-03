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

type HallService struct {
	db *gorm.DB
}

func NewHallService(db *gorm.DB) *HallService {
	return &HallService{db: db}
}

func (h *HallService) GetHall(ctx context.Context, hallID uint) (*entity.CinemaHall, error) {
	var hall entity.CinemaHall

	result := h.db.
		WithContext(ctx).
		First(&hall, hallID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrHallNotFound
		}
		return nil, result.Error
	}

	return &hall, nil
}

func (h *HallService) GetHallsPaginated(ctx context.Context, pageStr string, limitStr string) (*entity.PaginatedResponse, error) {
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

	halls := []entity.CinemaHall{}
	result := h.db.
		WithContext(ctx).
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&halls)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrHallNotFound
		}
		return nil, result.Error
	}

	var totalItems int64
	h.db.WithContext(ctx).Model(&entity.CinemaHall{}).Count(&totalItems)
	totalPages := (int(totalItems) + limit - 1) / limit

	response := entity.PaginatedResponse{
		Items:      halls,
		Page:       page,
		Limit:      limit,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
	}

	return &response, nil
}

func (h *HallService) CreateHall(ctx context.Context, hall entity.CinemaHall) (*entity.CinemaHall, error) {
	tx := h.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("hall create: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	if err := tx.Create(&hall).Error; err != nil {
		slog.Error("hall create: insert", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("hall create: commit", "err", err)
		return nil, ErrInternalServer
	}

	return &hall, nil
}

func (h *HallService) UpdateHall(ctx context.Context, hallID uint, changedHall entity.CinemaHall) (*entity.CinemaHall, error) {
	tx := h.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("hall update: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	hall := entity.CinemaHall{}
	result := tx.First(&hall, hallID)
	if result.Error != nil {
		slog.Error("hall update: hall not found", "err", result.Error)
		return nil, ErrHallNotFound
	}

	if err := tx.Save(&changedHall).Error; err != nil {
		slog.Error("hall update: insert", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("hall update: commit", "err", err)
		return nil, ErrInternalServer
	}

	return &changedHall, nil
}

func (h *HallService) DeleteHall(ctx context.Context, hallID uint) error {
	hall := entity.CinemaHall{}
	result := h.db.WithContext(ctx).First(&hall, hallID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrHallNotFound
		}
		return result.Error
	}

	tx := h.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("hall delete: begin tx", "err", tx.Error)
		return ErrInternalServer
	}
	defer tx.Rollback()

	if err := tx.Delete(&entity.CinemaHall{}, hallID).Error; err != nil {
		slog.Error("hall delete: insert", "err", err)
		return ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("hall delete: commit", "err", err)
		return ErrInternalServer
	}

	return nil
}
