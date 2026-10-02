package model

import "gorm.io/gorm"

type SApiPermissionModel struct {
	db *gorm.DB
}

func NewSApiPermissionModel(db *gorm.DB) *SApiPermissionModel {
	return &SApiPermissionModel{db: db}
}
