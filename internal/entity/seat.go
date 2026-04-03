package entity

import "gorm.io/gorm"

// Seat is one place in the hall grid for a specific session.
// UserID is nil until someone buys that seat.
type Seat struct {
	gorm.Model

	X         int     `gorm:"uniqueIndex:idx_session_seat_xy"`
	Y         int     `gorm:"uniqueIndex:idx_session_seat_xy"`
	SessionID uint    `gorm:"uniqueIndex:idx_session_seat_xy;not null"`
	Session   Session `gorm:"foreignKey:SessionID"`

	UserID *uint
	User   User `gorm:"foreignKey:UserID"`
}
