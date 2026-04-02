package services

import "gorm.io/gorm"

type MovieService struct {
	db *gorm.DB
}

func NewMovieService(db *gorm.DB) *MovieService {
	return &MovieService{db: db}
}
