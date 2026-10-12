// internal/db/db.go
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

const Dir = "./db"

//go:embed schema.sql
var schema string

// DB wraps an open SQLite database.
type DB struct {
	*sql.DB
	Name string
}

func dbPath(name string) string {
	return filepath.Join(Dir, name+".sqlite")
}

// List returns the names (without extension) of existing databases.
func List() ([]string, error) {
	entries, err := os.ReadDir(Dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sqlite") {
			names = append(names, strings.TrimSuffix(e.Name(), ".sqlite"))
		}
	}
	sort.Strings(names)
	return names, nil
}

func Exists(name string) bool {
	_, err := os.Stat(dbPath(name))
	return err == nil
}

// ValidateName checks a proposed DB name, returning a reason if invalid.
func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return fmt.Errorf("name must contain only alphanumeric characters")
		}
	}
	if Exists(name) {
		return fmt.Errorf("a database named %q already exists", name)
	}
	return nil
}

// Create creates a new database from the schema and seeds it with dummy records.
func Create(name string) (*DB, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}
	db, err := Open(name)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		os.Remove(dbPath(name))
		return nil, fmt.Errorf("failed to apply schema: %w", err)
	}
	if err := seed(db); err != nil {
		db.Close()
		os.Remove(dbPath(name))
		return nil, fmt.Errorf("failed to seed records: %w", err)
	}
	return db, nil
}

// Open opens an existing database.
func Open(name string) (*DB, error) {
	d, err := sql.Open("sqlite", dbPath(name))
	if err != nil {
		return nil, err
	}
	if err := d.Ping(); err != nil {
		d.Close()
		return nil, err
	}
	return &DB{DB: d, Name: name}, nil
}

func seed(db *DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO records (name, value) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for i := 1; i <= 15; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("dummy-%02d", i), i*10); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// RecordCount returns the number of rows in records.
func (d *DB) RecordCount() (int, error) {
	var n int
	err := d.QueryRow("SELECT COUNT(*) FROM records").Scan(&n)
	return n, err
}

// FirstRows returns up to limit rows from records as (rowid, col values) tuples.
func (d *DB) FirstRows(limit int) ([]string, [][]string, error) {
	rows, err := d.Query("SELECT * FROM records ORDER BY rowid LIMIT ?", limit)
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
			switch t := v.(type) {
			case nil:
				row[i] = "NULL"
			case []byte:
				row[i] = string(t)
			default:
				row[i] = fmt.Sprintf("%v", t)
			}
		}
		out = append(out, row)
	}
	return cols, out, rows.Err()
}
