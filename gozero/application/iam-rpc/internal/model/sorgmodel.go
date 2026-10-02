package model

import "gorm.io/gorm"

type SOrgModel struct {
	db *gorm.DB
}

func NewSOrgModel(db *gorm.DB) *SOrgModel {
	return &SOrgModel{db: db}
}
