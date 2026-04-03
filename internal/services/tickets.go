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

var ErrTicketNotFound = errors.New("ticket not found")

type TicketService struct {
	db *gorm.DB
}

func NewTicketService(db *gorm.DB) *TicketService {
	return &TicketService{db: db}
}

func (t *TicketService) GetTicket(ctx context.Context, userId uint, ticketID uint) (*entity.Ticket, error) {
	var ticket entity.Ticket

	result := t.db.
		WithContext(ctx).
		Where("user_id = ?", userId).
		Preload("Session").
		Preload("Seat").
		First(&ticket, ticketID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, result.Error
	}

	return &ticket, nil
}

func (t *TicketService) GetTicketsPaginated(ctx context.Context, userId uint, pageStr string, limitStr string) (*entity.PaginatedResponse, error) {
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

	// get movies
	tickets := []entity.Ticket{}
	result := t.db.
		WithContext(ctx).
		Where("user_id = ?", userId).
		Limit(limit).
		Offset((page - 1) * limit).
		Preload("Session").
		Preload("Seat").
		Find(&tickets)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, result.Error
	}

	var totalItems int64
	t.db.WithContext(ctx).Model(&entity.Ticket{}).Count(&totalItems)
	totalPages := (int(totalItems) + limit - 1) / limit

	response := entity.PaginatedResponse{
		Items:      tickets,
		Page:       page,
		Limit:      limit,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
	}

	return &response, nil
}

func (t *TicketService) BuyTicket(ctx context.Context, userId uint, ticket entity.Ticket) (*entity.Ticket, error) {
	tx := t.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("ticket buy: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	ticket.UserID = userId
	ticket.PurchasePrice = ticket.Session.Price

	if err := tx.Create(&ticket).Error; err != nil {
		slog.Error("ticket buy: insert", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("ticket buy: commit", "err", err)
		return nil, ErrInternalServer
	}

	return &ticket, nil
}

func (t *TicketService) CancelTicket(ctx context.Context, userId uint, ticketId uint) error {
	tx := t.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("ticket cancel: begin tx", "err", tx.Error)
		return ErrInternalServer
	}
	defer tx.Rollback()

	// check if ticket exists
	ticket := entity.Ticket{}
	result := tx.
		WithContext(ctx).
		Where("user_id = ?", userId).
		Where("id = ?", ticketId).
		First(&ticket)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrTicketNotFound
		}
		return result.Error
	}

	// delete ticket
	if err := tx.Delete(&entity.Ticket{}, ticketId).Error; err != nil {
		slog.Error("ticket cancel: insert", "err", err)
		return ErrInternalServer
	}

	// cancel reservation in seat
	seat := entity.Seat{}
	result = tx.
		WithContext(ctx).
		Where("user_id = ?", userId).
		Where("id = ?", ticket.SeatID).
		First(&seat)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrTicketNotFound
		}
		return result.Error
	}

	seat.UserID = 0

	if err := tx.Save(&seat).Error; err != nil {
		slog.Error("ticket cancel: save seat", "err", err)
		return ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("ticket cancel: commit", "err", err)
		return ErrInternalServer
	}

	return nil
}
