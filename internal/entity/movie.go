package entity

import (
	"time"

	"gorm.io/gorm"
)

type Movie struct {
	gorm.Model

	Name        string    `json:"name"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"release_date"`
	Duration    int       `json:"duration"` // Movie duration in minutes

	Sessions []Session `gorm:"foreignKey:MovieID"`
	Users    []User    `gorm:"many2many:user_watchlists"`
}
