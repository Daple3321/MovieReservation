package entity

import (
	"time"

	"gorm.io/gorm"
)

// Session is one screening of a movie.
type Session struct {
	gorm.Model

	HallID     uint       `gorm:"not null"`
	CinemaHall CinemaHall `gorm:"foreignKey:HallID"`
	MovieID    uint       `gorm:"not null"`
	Movie      Movie      `gorm:"foreignKey:MovieID"`

	Price     float64   `gorm:"not null;default:0.00"`
	Seats     []Seat    `gorm:"foreignKey:SessionID"`
	StartTime time.Time `gorm:"not null"`

	Tickets []Ticket `gorm:"foreignKey:SessionID"`
}
