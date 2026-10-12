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

const (
	Dir = "./db"
	Ext = ".db"
)

//go:embed schema.sql
var schema string

var nameRe = regexp.MustCompile(`^[A-Za-z0-9]+$`)

type DB struct {
	*sql.DB
	Name string
}

func EnsureDir() error {
	return os.MkdirAll(Dir, 0o755)
}

func path(name string) string {
	return filepath.Join(Dir, name+Ext)
}

func List() ([]string, error) {
	entries, err := os.ReadDir(Dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != Ext {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), Ext))
	}
	sort.Strings(names)
	return names, nil
}

func ValidateName(name string, existing []string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if !nameRe.MatchString(name) {
		return errors.New("name must contain only alphanumeric characters")
	}
	for _, e := range existing {
		if strings.EqualFold(e, name) {
			return fmt.Errorf("a database named %q already exists", e)
		}
	}
	return nil
}

func Open(name string) (*DB, error) {
	conn, err := sql.Open("sqlite", path(name))
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return &DB{DB: conn, Name: name}, nil
}

func Create(name string) (*DB, error) {
	existing, err := List()
	if err != nil {
		return nil, err
	}
	if err := ValidateName(name, existing); err != nil {
		return nil, err
	}
	d, err := Open(name)
	if err != nil {
		return nil, err
	}
	if _, err := d.Exec(schema); err != nil {
		d.Close()
		os.Remove(path(name))
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := seed(d.DB, 15); err != nil {
		d.Close()
		os.Remove(path(name))
		return nil, fmt.Errorf("seed: %w", err)
	}
	return d, nil
}

func (d *DB) Count() (int, error) {
	var n int
	err := d.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&n)
	return n, err
}

// Preview returns column names and up to limit rows (as strings) from records.
func (d *DB) Preview(limit int) ([]string, [][]string, error) {
	rows, err := d.Query(`SELECT rowid, * FROM records ORDER BY rowid LIMIT ?`, limit)
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
		rec := make([]string, len(cols))
		for i, v := range vals {
			rec[i] = toString(v)
		}
		out = append(out, rec)
	}
	return cols, out, rows.Err()
}

func toString(v any) string {
	switch x := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(x)
	default:
		return fmt.Sprint(x)
	}
}
