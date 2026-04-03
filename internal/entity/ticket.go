package entity

import "gorm.io/gorm"

type Ticket struct {
	gorm.Model

	SessionID uint    `gorm:"not null"`
	Session   Session `gorm:"foreignKey:SessionID"`

	SeatID uint `gorm:"not null"`
	Seat   Seat `gorm:"foreignKey:SeatID"`

	UserID uint `gorm:"not null"`
	User   User `gorm:"foreignKey:UserID"`

	PurchasePrice float64 `gorm:"type:decimal(10,2);not null;default:0.00"`
}
