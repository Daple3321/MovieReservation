package entity

import (
	"time"

	"gorm.io/gorm"
)

// Session is one screening of a movie.
type Session struct {
	gorm.Model

	HallID     uint
	CinemaHall CinemaHall `gorm:"foreignKey:HallID"`
	MovieID    uint
	Movie      Movie `gorm:"foreignKey:MovieID"`

	Price          float64
	SeatsAvailable int
	StartTime      time.Time

	Tickets []Ticket `gorm:"foreignKey:SessionID"`
}
