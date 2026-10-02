package {{.pkg}}

import "gorm.io/gorm"

type {{.upperStartCamelObject}}Model struct {
	db *gorm.DB
}

func New{{.upperStartCamelObject}}Model(db *gorm.DB) *{{.upperStartCamelObject}}Model {
	return &{{.upperStartCamelObject}}Model{db: db}
}
