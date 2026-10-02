package model

import "gorm.io/gorm"

type SLoginLogModel struct {
	db *gorm.DB
}

func NewSLoginLogModel(db *gorm.DB) *SLoginLogModel {
	return &SLoginLogModel{db: db}
}
