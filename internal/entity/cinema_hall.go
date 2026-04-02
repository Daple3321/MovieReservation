package entity

import "gorm.io/gorm"

type CinemaHall struct {
	gorm.Model

	Seats int

	Sessions []Session `gorm:"foreignKey:HallID"`
}
