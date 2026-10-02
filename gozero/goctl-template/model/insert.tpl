func (m *{{.upperStartCamelObject}}Model) Insert(ctx context.Context, data *{{.upperStartCamelObject}}) (sql.Result, error) {
	result := gorm.WithResult()
	err := gorm.G[{{.upperStartCamelObject}}](m.db, result).Create(ctx, data)
	return result.Result, err
}
