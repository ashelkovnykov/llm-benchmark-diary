package db

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

const (
	Ext              = ".db"
	RecordsTable     = "records"
	DummyRecordCount = 15
)

var validName = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// List returns the names (without extension) of all databases in dir.
func List(dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read db dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), Ext) {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))
	}
	sort.Strings(names)
	return names, nil
}

// ValidateName checks a proposed database name against the naming rules.
func ValidateName(name string, existing []string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if !validName.MatchString(name) {
		return errors.New("name may only contain alphanumeric characters (a-z, A-Z, 0-9)")
	}
	for _, e := range existing {
		if strings.EqualFold(e, name) {
			return fmt.Errorf("a database named %q already exists", e)
		}
	}
	return nil
}

func path(dir, name string) string {
	return filepath.Join(dir, name+Ext)
}

// Open opens an existing database.
func Open(dir, name string) (*sql.DB, error) {
	p := path(dir, name)
	if _, err := os.Stat(p); err != nil {
		return nil, fmt.Errorf("database %q: %w", name, err)
	}
	conn, err := sql.Open("sqlite", p)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// Create validates the name, creates a new database from the schema,
// and seeds it with dummy records.
func Create(dir, name string, existing []string) (conn *sql.DB, err error) {
	if err := ValidateName(name, existing); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	p := path(dir, name)
	if _, err := os.Stat(p); err == nil {
		return nil, fmt.Errorf("a database named %q already exists", name)
	}

	conn, err = sql.Open("sqlite", p)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			conn.Close()
			os.Remove(p)
			conn = nil
		}
	}()

	if _, err = conn.Exec(schema); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err = seed(conn, DummyRecordCount); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}
	return conn, nil
}

// Count returns the number of rows in the records table.
func Count(conn *sql.DB) (int, error) {
	var n int
	err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", quote(RecordsTable))).Scan(&n)
	return n, err
}

// Preview returns column names and up to limit rows from the records table.
func Preview(conn *sql.DB, limit int) ([]string, [][]string, error) {
	rows, err := conn.Query(
		fmt.Sprintf("SELECT * FROM %s ORDER BY rowid LIMIT ?", quote(RecordsTable)), limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var out [][]string
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			row[i] = format(v)
		}
		out = append(out, row)
	}
	return cols, out, rows.Err()
}

func format(v any) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	default:
		return fmt.Sprint(t)
	}
}

func quote(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}
