package model

import "gorm.io/gorm"

type SOperLogModel struct {
	db *gorm.DB
}

func NewSOperLogModel(db *gorm.DB) *SOperLogModel {
	return &SOperLogModel{db: db}
}
