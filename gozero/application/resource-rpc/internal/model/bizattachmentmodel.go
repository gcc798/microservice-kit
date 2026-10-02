package model

import "gorm.io/gorm"

type BizAttachmentModel struct {
	db *gorm.DB
}

func NewBizAttachmentModel(db *gorm.DB) *BizAttachmentModel {
	return &BizAttachmentModel{db: db}
}
