package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed migrations/001_create_messages.sql
var schema string

func Migrate(db *sql.DB) error {
	for _, stmt := range strings.Split(schema, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("execute migration: %w", err)
		}
	}
	return nil
}
