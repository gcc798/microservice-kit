package model

import "gorm.io/gorm"

type SConfigModel struct {
	db *gorm.DB
}

func NewSConfigModel(db *gorm.DB) *SConfigModel {
	return &SConfigModel{db: db}
}
