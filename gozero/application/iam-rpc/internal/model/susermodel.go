package model

import "gorm.io/gorm"

type SUserModel struct {
	db *gorm.DB
}

func NewSUserModel(db *gorm.DB) *SUserModel {
	return &SUserModel{db: db}
}
