package model

import "gorm.io/gorm"

type MMenuApiPermissionModel struct {
	db *gorm.DB
}

func NewMMenuApiPermissionModel(db *gorm.DB) *MMenuApiPermissionModel {
	return &MMenuApiPermissionModel{db: db}
}
