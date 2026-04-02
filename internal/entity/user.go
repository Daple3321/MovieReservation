package entity

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Username     string `gorm:"unique"`
	PasswordHash string
	Role         string

	Watchlist []Movie  `gorm:"many2many:user_watchlists"`
	Tickets   []Ticket `gorm:"foreignKey:UserID"`
}

type UserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
