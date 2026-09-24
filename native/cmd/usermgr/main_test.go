package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestUserManagerParseOptionsAndValidateInput(t *testing.T) {
	opts, err := parseOptions([]string{
		"--operation=create",
		"--username=admin",
		"--nickname=管理员",
		"--role=super_admin",
		"--database-dsn=postgres-dsn",
	})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}
	if opts.operation != "create" || opts.username != "admin" || opts.databaseDSN != "postgres-dsn" {
		t.Fatalf("parseOptions() = %#v", opts)
	}
	if err := validateInput(opts, "strong-password"); err != nil {
		t.Fatalf("validateInput() error = %v", err)
	}
}

func TestUserManagerUsesDefaultDatabaseDSN(t *testing.T) {
	opts, err := parseOptions(nil)
	if err != nil {
		t.Fatal(err)
	}
	if opts.databaseDSN != defaultDatabaseDSN {
		t.Fatalf("databaseDSN = %q, want default", opts.databaseDSN)
	}
}

func TestValidateInputRejectsUnsafeInput(t *testing.T) {
	tests := []struct {
		name     string
		opts     options
		password string
	}{
		{name: "unknown operation", opts: options{operation: "delete", username: "admin", databaseDSN: defaultDatabaseDSN}, password: "strong-password"},
		{name: "missing username", opts: options{operation: "reset", databaseDSN: defaultDatabaseDSN}, password: "strong-password"},
		{name: "short password", opts: options{operation: "reset", username: "admin", databaseDSN: defaultDatabaseDSN}, password: "short"},
		{name: "missing database DSN", opts: options{operation: "reset", username: "admin"}, password: "strong-password"},
		{name: "missing role", opts: options{operation: "create", username: "admin", nickname: "管理员", databaseDSN: defaultDatabaseDSN}, password: "strong-password"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateInput(tt.opts, tt.password); err == nil {
				t.Fatal("validateInput() expected error")
			}
		})
	}
}

func TestUserManagerModelsGenerateIDs(t *testing.T) {
	for _, model := range []any{user{}, userRole{}} {
		field := reflect.TypeOf(model).Field(0)
		if field.Tag.Get("autogen") != "int64" || !strings.Contains(field.Tag.Get("gorm"), "primaryKey") || !strings.Contains(field.Tag.Get("gorm"), "autoIncrement:false") {
			t.Fatalf("%T ID tags = %q", model, field.Tag)
		}
	}
}
