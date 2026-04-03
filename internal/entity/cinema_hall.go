package entity

import "gorm.io/gorm"

// CinemaHall acts like a template to populate session seats for a new session.
type CinemaHall struct {
	gorm.Model

	Width  int `gorm:"not null"`
	Height int `gorm:"not null"`

	Sessions []Session `gorm:"foreignKey:HallID"`
}
