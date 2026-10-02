package model

import "gorm.io/gorm"

type SRoleModel struct {
	db *gorm.DB
}

func NewSRoleModel(db *gorm.DB) *SRoleModel {
	return &SRoleModel{db: db}
}
