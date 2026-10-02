package database

import (
	"fmt"
	"reflect"

	"github.com/gcc798/microservice-kit/internal/idgen"
	"gorm.io/gorm"
)

// IDGenPlugin fills integer primary keys marked with autogen:"int64".
type IDGenPlugin struct{}

func (*IDGenPlugin) Name() string { return "idgen" }

func (p *IDGenPlugin) Initialize(db *gorm.DB) error {
	return db.Callback().Create().Before("gorm:create").Register("idgen:before_create", p.beforeCreate)
}

func (p *IDGenPlugin) beforeCreate(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}

	value := db.Statement.ReflectValue
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		for i := range value.Len() {
			p.setID(db, indirect(value.Index(i)))
		}
		return
	}
	p.setID(db, indirect(value))
}

func (*IDGenPlugin) setID(db *gorm.DB, value reflect.Value) {
	for _, field := range db.Statement.Schema.Fields {
		if field.Tag.Get("autogen") != "int64" {
			continue
		}
		_, zero := field.ValueOf(db.Statement.Context, value)
		if !zero {
			continue
		}
		id, err := idgen.NextID()
		if err != nil {
			db.AddError(fmt.Errorf("generate ID for %s: %w", field.Name, err))
			return
		}
		if err := field.Set(db.Statement.Context, value, id); err != nil {
			db.AddError(fmt.Errorf("set ID for %s: %w", field.Name, err))
			return
		}
	}
}

func indirect(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	return value
}
