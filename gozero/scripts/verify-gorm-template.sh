#!/bin/sh

set -eu

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
template_dir="$root_dir/goctl-template"
goctl_command=${GOCTL:-go tool goctl}
test_dsn=${GOCTL_GORM_TEST_DSN:-postgres://postgres:post123@127.0.0.1:5433/postgres?sslmode=disable}
schema="goctl_gorm_test_$$"
output_dir=$(mktemp -d /tmp/goctl-gorm-template.XXXXXX)

cleanup() {
	psql "$test_dsn" -v ON_ERROR_STOP=1 -c "DROP SCHEMA IF EXISTS $schema CASCADE" >/dev/null
	rm -rf "$output_dir"
}
trap cleanup EXIT INT TERM

psql "$test_dsn" -v ON_ERROR_STOP=1 <<SQL >/dev/null
CREATE SCHEMA $schema;
CREATE TABLE $schema.template_users (
    id bigint PRIMARY KEY,
    user_name varchar(64) NOT NULL UNIQUE,
    display_name varchar(128),
    status smallint NOT NULL DEFAULT 0,
    created_time timestamptz NOT NULL DEFAULT now(),
    updated_time timestamptz NOT NULL DEFAULT now()
);
SQL

${goctl_command} model pg datasource \
	--url="$test_dsn" \
	--schema="$schema" \
	--table=template_users \
	--dir="$output_dir/model" \
	--strict \
	--home="$template_dir"

if rg -n 'go-zero/core/stores/(sqlx|sqlc)' "$output_dir"; then
	echo "generated model unexpectedly depends on go-zero SQLX" >&2
	exit 1
fi
if ! rg -q 'primaryKey;autoIncrement:false.*autogen:"int64"' "$output_dir/model"; then
	echo "generated model is missing the distributed integer primary-key tags" >&2
	exit 1
fi

cp "$root_dir/testdata/gorm_model_runtime_test.go" "$output_dir/model/gorm_model_runtime_test.go"
(
	cd "$output_dir"
	go mod init example.com/goctl-gorm-template-test
	go get gorm.io/driver/postgres@v1.6.3 gorm.io/gorm@v1.31.2
	GOCTL_GORM_TEST_DSN="${test_dsn}&search_path=$schema" go test ./...
)

(cd "$root_dir" && GOZERO_DATABASE_TEST_DSN="$test_dsn" go test ./internal/database -run TestIDGenPluginPostgres)
