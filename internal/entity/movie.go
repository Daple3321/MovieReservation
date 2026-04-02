package entity

import (
	"time"

	"gorm.io/gorm"
)

type Movie struct {
	gorm.Model

	Name        string
	Description string
	ReleaseDate time.Time
	Duration    int

	Sessions []Session `gorm:"foreignKey:MovieID"`
	Users    []User    `gorm:"many2many:user_watchlists"`
}
