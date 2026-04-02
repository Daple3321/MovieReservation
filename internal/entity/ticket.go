package entity

import "gorm.io/gorm"

type Ticket struct {
	gorm.Model
	SessionID     uint
	Session       Session
	UserID        uint
	User          User
	PurchasePrice float64
}
