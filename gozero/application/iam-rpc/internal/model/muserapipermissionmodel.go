package model

import "gorm.io/gorm"

type MUserApiPermissionModel struct {
	db *gorm.DB
}

func NewMUserApiPermissionModel(db *gorm.DB) *MUserApiPermissionModel {
	return &MUserApiPermissionModel{db: db}
}
