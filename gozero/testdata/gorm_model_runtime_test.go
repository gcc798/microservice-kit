package model

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGeneratedModelCRUD(t *testing.T) {
	db, err := gorm.Open(postgres.Open(os.Getenv("GOCTL_GORM_TEST_DSN")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	users := NewTemplateUsersModel(db)
	user := &TemplateUsers{Id: 42, UserName: "generated-user", Status: 0}
	if _, err := users.Insert(t.Context(), user); err != nil {
		t.Fatal(err)
	}
	if user.Id != 42 {
		t.Fatalf("generated insert changed the primary key to %d", user.Id)
	}

	found, err := users.FindOneByUserName(t.Context(), user.UserName)
	if err != nil {
		t.Fatal(err)
	}
	found.DisplayName = sql.NullString{String: "Generated User", Valid: true}
	found.Status = 1
	if err := users.Update(t.Context(), found); err != nil {
		t.Fatal(err)
	}

	found, err = users.FindOne(t.Context(), user.Id)
	if err != nil {
		t.Fatal(err)
	}
	if found.DisplayName.String != "Generated User" || found.Status != 1 {
		t.Fatalf("unexpected updated user: %#v", found)
	}

	if err := users.Delete(t.Context(), user.Id); err != nil {
		t.Fatal(err)
	}
	if _, err := users.FindOne(t.Context(), user.Id); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found, got %v", err)
	}
}
