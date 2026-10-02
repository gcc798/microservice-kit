package model

import "gorm.io/gorm"

type MRoleApiPermissionModel struct {
	db *gorm.DB
}

func NewMRoleApiPermissionModel(db *gorm.DB) *MRoleApiPermissionModel {
	return &MRoleApiPermissionModel{db: db}
}
