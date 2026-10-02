package model

import "gorm.io/gorm"

type SAuthClientModel struct {
	db *gorm.DB
}

func NewSAuthClientModel(db *gorm.DB) *SAuthClientModel {
	return &SAuthClientModel{db: db}
}
