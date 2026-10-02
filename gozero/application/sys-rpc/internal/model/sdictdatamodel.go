package model

import "gorm.io/gorm"

type SDictDataModel struct {
	db *gorm.DB
}

func NewSDictDataModel(db *gorm.DB) *SDictDataModel {
	return &SDictDataModel{db: db}
}
