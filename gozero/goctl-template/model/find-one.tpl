func (m *{{.upperStartCamelObject}}Model) FindOne(ctx context.Context, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) (*{{.upperStartCamelObject}}, error) {
	data, err := gorm.G[{{.upperStartCamelObject}}](m.db).
		Where("{{.originalPrimaryKey}} = ?", {{.lowerStartCamelPrimaryKey}}).
		First(ctx)
	return &data, err
}
