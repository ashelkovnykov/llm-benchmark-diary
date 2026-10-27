package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

const Dir = "./db"
const Ext = ".db"

func EnsureDir() error {
	return os.MkdirAll(Dir, 0o755)
}

func List() ([]string, error) {
	if err := EnsureDir(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(Dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), Ext) {
			names = append(names, strings.TrimSuffix(e.Name(), Ext))
		}
	}
	sort.Strings(names)
	return names, nil
}

func ValidateName(name string, existing []string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return fmt.Errorf("name %q contains invalid character %q: only alphanumeric characters allowed", name, r)
		}
	}
	for _, e := range existing {
		if e == name {
			return fmt.Errorf("a database named %q already exists", name)
		}
	}
	return nil
}

func Path(name string) string {
	return filepath.Join(Dir, name+Ext)
}

func Create(name string) (*sql.DB, error) {
	if err := EnsureDir(); err != nil {
		return nil, err
	}
	conn, err := Open(name)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	for i := 1; i <= 15; i++ {
		if _, err := conn.Exec(
			`INSERT INTO records (title, value) VALUES (?, ?)`,
			fmt.Sprintf("dummy-%02d", i), i*10,
		); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return conn, nil
}

func Open(name string) (*sql.DB, error) {
	if err := EnsureDir(); err != nil {
		return nil, err
	}
	return sql.Open("sqlite", Path(name))
}

func Count(conn *sql.DB) (int, error) {
	var n int
	err := conn.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&n)
	return n, err
}

func FirstRows(conn *sql.DB, limit int) ([]string, error) {
	rows, err := conn.Query(`SELECT * FROM records ORDER BY rowid LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var out []string
	out = append(out, strings.Join(cols, " | "))
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		parts := make([]string, len(cols))
		for i, v := range vals {
			switch t := v.(type) {
			case []byte:
				parts[i] = string(t)
			default:
				parts[i] = fmt.Sprintf("%v", t)
			}
		}
		out = append(out, strings.Join(parts, " | "))
	}
	return out, rows.Err()
}
