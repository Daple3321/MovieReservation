package entity

import "gorm.io/gorm"

// Seat tied to a specific session and a specific user.
type Seat struct {
	gorm.Model

	X         int     `gorm:"uniqueIndex:idx_seat"`
	Y         int     `gorm:"uniqueIndex:idx_seat"`
	SessionID uint    `gorm:"uniqueIndex:idx_seat;not null"`
	Session   Session `gorm:"foreignKey:SessionID"`
	UserID    uint    `gorm:"uniqueIndex:idx_seat"`
	User      User    `gorm:"foreignKey:UserID"`
}
