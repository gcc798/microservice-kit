package database

import (
	"fmt"
	"os"
	"testing"
	"time"
)

type idRecord struct {
	ID int64 `gorm:"column:id;primaryKey;autoIncrement:false" autogen:"int64"`
}

func TestIDGenPluginPostgres(t *testing.T) {
	dsn := os.Getenv("GOZERO_DATABASE_TEST_DSN")
	if dsn == "" {
		t.Skip("GOZERO_DATABASE_TEST_DSN is not set")
	}

	db, err := Open(Config{DSN: dsn})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(db) })

	schema := fmt.Sprintf("gorm_idgen_test_%d", time.Now().UnixNano())
	if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE").Error })

	table := schema + ".records"
	if err := db.Exec("CREATE TABLE " + table + " (id bigint PRIMARY KEY)").Error; err != nil {
		t.Fatal(err)
	}
	record := idRecord{}
	if err := db.WithContext(t.Context()).Table(table).Create(&record).Error; err != nil {
		t.Fatal(err)
	}
	if record.ID == 0 {
		t.Fatal("ID generator did not populate the primary key")
	}
}
