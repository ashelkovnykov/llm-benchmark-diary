package db

import (
	"os"
	"path/filepath"
	"strings"
)

var schema = `
CREATE TABLE records (
    rowid INTEGER PRIMARY KEY,
    name TEXT,
    description TEXT,
    value INTEGER
);

CREATE INDEX idx_records_rowid ON records(rowid);
`

func GetSchema() string {
	return strings.TrimSpace(schema)
}

func GetSchemaPath() string {
	exe, _ := os.Executable()
	base := filepath.Dir(exe)
	if base == "." {
		base, _ = os.Getwd()
	}
	return filepath.Join(base, "internal", "db", "schema.sql")
}
