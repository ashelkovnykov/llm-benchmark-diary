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
	"time"

	_ "modernc.org/sqlite"
)

const (
	Dir          = "./db"
	Ext          = ".db"
	DummyRecords = 15
	PreviewRows  = 10
)

//go:embed schema.sql
var schema string

var nameRe = regexp.MustCompile(`^[A-Za-z0-9]+$`)

var (
	ErrEmptyName   = errors.New("name cannot be empty")
	ErrInvalidName = errors.New("name may only contain letters and digits")
	ErrNameExists  = errors.New("a database with this name already exists")
)

// Path returns the file path for a database with the given name.
func Path(name string) string {
	return filepath.Join(Dir, name+Ext)
}

// ValidateName checks that a name is non-empty, alphanumeric, and unused.
func ValidateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if !nameRe.MatchString(name) {
		return ErrInvalidName
	}
	if _, err := os.Stat(Path(name)); err == nil {
		return ErrNameExists
	}
	return nil
}

// List returns the names (without extension) of all databases in Dir.
func List() ([]string, error) {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating %s: %w", Dir, err)
	}
	entries, err := os.ReadDir(Dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", Dir, err)
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

// Create validates the name, then creates and seeds a new database.
func Create(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", Dir, err)
	}

	path := Path(name)
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	if err := seed(conn); err != nil {
		conn.Close()
		os.Remove(path)
		return err
	}
	return conn.Close()
}

func seed(conn *sql.DB) error {
	if _, err := conn.Exec(schema); err != nil {
		return fmt.Errorf("applying schema: %w", err)
	}

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO records (name, value, created_at) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for i := 1; i <= DummyRecords; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("record-%02d", i), float64(i)*1.5, now); err != nil {
			return fmt.Errorf("seeding records: %w", err)
		}
	}
	return tx.Commit()
}

// Open opens an existing database by name.
func Open(name string) (*sql.DB, error) {
	path := Path(name)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return sql.Open("sqlite", path)
}

// Count returns the number of rows in the records table.
func Count(conn *sql.DB) (int, error) {
	var n int
	err := conn.QueryRow("SELECT COUNT(*) FROM records").Scan(&n)
	return n, err
}

// Preview returns the column names and up to limit rows from records.
func Preview(conn *sql.DB, limit int) ([]string, [][]string, error) {
	rows, err := conn.Query("SELECT * FROM records ORDER BY rowid LIMIT ?", limit)
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
			switch x := v.(type) {
			case nil:
				row[i] = "NULL"
			case []byte:
				row[i] = string(x)
			default:
				row[i] = fmt.Sprint(x)
			}
		}
		out = append(out, row)
	}
	return cols, out, rows.Err()
}
