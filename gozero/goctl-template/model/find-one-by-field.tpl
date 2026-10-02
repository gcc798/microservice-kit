func (m *{{.upperStartCamelObject}}Model) FindOneBy{{.upperField}}(ctx context.Context, {{.in}}) (*{{.upperStartCamelObject}}, error) {
	data, err := gorm.G[{{.upperStartCamelObject}}](m.db).
		Where("{{.originalField}}", {{.lowerStartCamelField}}).
		First(ctx)
	return &data, err
}
