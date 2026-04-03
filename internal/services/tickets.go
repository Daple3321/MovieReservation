package services

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Daple3321/MovieReservation/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrTicketNotFound       = errors.New("ticket not found")
	ErrSeatNotFound         = errors.New("seat not found")
	ErrSeatNotAvailable     = errors.New("seat is not available")
	ErrSeatSessionMismatch  = errors.New("seat does not belong to this session")
	ErrSessionMismatch      = errors.New("session not found for this ticket")
	ErrInvalidTicketRequest = errors.New("session_id and seat_id are required")
)

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
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

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
		return nil, result.Error
	}

	var totalItems int64
	if err := t.db.WithContext(ctx).
		Model(&entity.Ticket{}).
		Where("user_id = ?", userId).
		Count(&totalItems).Error; err != nil {
		return nil, err
	}

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

func (t *TicketService) BuyTicket(ctx context.Context, userId uint, in entity.Ticket) (*entity.Ticket, error) {
	if in.SessionID == 0 || in.SeatID == 0 {
		return nil, ErrInvalidTicketRequest
	}

	tx := t.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("ticket buy: begin tx", "err", tx.Error)
		return nil, ErrInternalServer
	}
	defer tx.Rollback()

	var session entity.Session
	if err := tx.First(&session, in.SessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionMismatch
		}
		slog.Error("ticket buy: load session", "err", err)
		return nil, ErrInternalServer
	}

	var seat entity.Seat
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seat, in.SeatID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSeatNotFound
		}
		slog.Error("ticket buy: load seat", "err", err)
		return nil, ErrInternalServer
	}

	if seat.SessionID != in.SessionID {
		return nil, ErrSeatSessionMismatch
	}

	if seat.UserID != nil {
		return nil, ErrSeatNotAvailable
	}

	uid := userId
	seat.UserID = &uid
	if err := tx.Save(&seat).Error; err != nil {
		slog.Error("ticket buy: assign seat", "err", err)
		return nil, ErrInternalServer
	}

	newTicket := entity.Ticket{
		SessionID:     in.SessionID,
		SeatID:        in.SeatID,
		UserID:        userId,
		PurchasePrice: session.Price,
	}

	if err := tx.Create(&newTicket).Error; err != nil {
		slog.Error("ticket buy: insert ticket", "err", err)
		return nil, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("ticket buy: commit", "err", err)
		return nil, ErrInternalServer
	}

	if err := t.db.WithContext(ctx).
		Preload("Session").
		Preload("Seat").
		First(&newTicket, newTicket.ID).Error; err != nil {
		return nil, err
	}

	return &newTicket, nil
}

func (t *TicketService) CancelTicket(ctx context.Context, userId uint, ticketId uint) error {
	tx := t.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("ticket cancel: begin tx", "err", tx.Error)
		return ErrInternalServer
	}
	defer tx.Rollback()

	ticket := entity.Ticket{}
	if err := tx.Where("user_id = ?", userId).Where("id = ?", ticketId).First(&ticket).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTicketNotFound
		}
		return err
	}

	var seat entity.Seat
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seat, ticket.SeatID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSeatNotFound
		}
		return err
	}

	if seat.UserID == nil || *seat.UserID != userId {
		return ErrSeatNotAvailable
	}

	if err := tx.Model(&entity.Seat{}).Where("id = ?", ticket.SeatID).Update("user_id", nil).Error; err != nil {
		slog.Error("ticket cancel: clear seat", "err", err)
		return ErrInternalServer
	}

	if err := tx.Delete(&entity.Ticket{}, ticketId).Error; err != nil {
		slog.Error("ticket cancel: delete ticket", "err", err)
		return ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("ticket cancel: commit", "err", err)
		return ErrInternalServer
	}

	return nil
}
