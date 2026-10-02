package model

import "gorm.io/gorm"

type MRoleMenuModel struct {
	db *gorm.DB
}

func NewMRoleMenuModel(db *gorm.DB) *MRoleMenuModel {
	return &MRoleMenuModel{db: db}
}
