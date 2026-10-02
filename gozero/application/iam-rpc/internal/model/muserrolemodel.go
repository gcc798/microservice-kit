package model

import "gorm.io/gorm"

type MUserRoleModel struct {
	db *gorm.DB
}

func NewMUserRoleModel(db *gorm.DB) *MUserRoleModel {
	return &MUserRoleModel{db: db}
}
