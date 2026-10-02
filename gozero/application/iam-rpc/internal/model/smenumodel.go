package model

import "gorm.io/gorm"

type SMenuModel struct {
	db *gorm.DB
}

func NewSMenuModel(db *gorm.DB) *SMenuModel {
	return &SMenuModel{db: db}
}
