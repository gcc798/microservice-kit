func (m *{{.upperStartCamelObject}}Model) Update(ctx context.Context, data *{{.upperStartCamelObject}}) error {
	return m.db.WithContext(ctx).
		Model(&{{.upperStartCamelObject}}{}).
		Where("{{.originalPrimaryKey}} = ?", data.{{.upperStartCamelPrimaryKey}}).
		Select("*").
		Omit("{{.originalPrimaryKey}}", "create_at", "created_at", "create_time", "update_at", "updated_at", "update_time").
		Updates(data).Error
}
