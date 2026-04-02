package handlers

import "gorm.io/gorm"

type SessionHandler struct {
	db *gorm.DB
}

func NewSessionHandler(db *gorm.DB) SessionHandler {
	return SessionHandler{db: db}
}
