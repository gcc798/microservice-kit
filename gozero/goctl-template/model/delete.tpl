func (m *{{.upperStartCamelObject}}Model) Delete(ctx context.Context, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) error {
	_, err := gorm.G[{{.upperStartCamelObject}}](m.db).
		Where("{{.originalPrimaryKey}} = ?", {{.lowerStartCamelPrimaryKey}}).
		Delete(ctx)
	return err
}
