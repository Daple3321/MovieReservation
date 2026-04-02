package entity

import (
	"time"

	"gorm.io/gorm"
)

// Session is one screening of a movie.
type Session struct {
	gorm.Model

	HallID     uint
	CinemaHall CinemaHall
	MovieID    uint
	Movie      Movie

	Price          float64
	SeatsAvailable int
	StartTime      time.Time

	Tickets []Ticket `gorm:"foreignKey:SessionID"`
}
