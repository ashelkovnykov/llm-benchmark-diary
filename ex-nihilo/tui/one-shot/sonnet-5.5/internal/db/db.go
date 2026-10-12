package db

import (
	_ "embed"

	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// Dir is the directory where databases are stored.
const Dir = "./db"

const ext = ".db"

const previewLimit = 10
const seedCount = 15

//go:embed schema.sql
var schema string

var nameRe = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// List returns the names (without extension) of existing databases in dir.
func List(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ext {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ext))
	}
	sort.Strings(names)
	return names, nil
}

// ValidateName checks a prospective database name.
func ValidateName(dir, name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if !nameRe.MatchString(name) {
		return errors.New("name may only contain letters and numbers")
	}
	existing, err := List(dir)
	if err != nil {
		return err
	}
	for _, e := range existing {
		if strings.EqualFold(e, name) {
			return fmt.Errorf("a database named %q already exists", e)
		}
	}
	return nil
}

func path(dir, name string) string {
	return filepath.Join(dir, name+ext)
}

// Open opens an existing database.
func Open(dir, name string) (*sql.DB, error) {
	p := path(dir, name)
	if _, err := os.Stat(p); err != nil {
		return nil, err
	}
	return connect(p)
}

// Create makes a new database from the embedded schema, seeds it with
// dummy records, and returns an open handle.
func Create(dir, name string) (*sql.DB, error) {
	if err := ValidateName(dir, name); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	p := path(dir, name)
	conn, err := connect(p)
	if err != nil {
		return nil, err
	}
	if err := initialize(conn); err != nil {
		conn.Close()
		os.Remove(p)
		return nil, err
	}
	return conn, nil
}

func connect(p string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", p)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func initialize(conn *sql.DB) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO records (name, category, value) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	categories := []string{"alpha", "beta", "gamma"}
	for i := 1; i <= seedCount; i++ {
		if _, err := stmt.Exec(
			fmt.Sprintf("Record %d", i),
			categories[i%len(categories)],
			float64(i)*1.5,
		); err != nil {
			return fmt.Errorf("seed: %w", err)
		}
	}
	return tx.Commit()
}

// Snapshot is a simple view of the records table.
type Snapshot struct {
	Total   int
	Columns []string
	Rows    [][]string
}

// Load returns the record count and the first rows of the records table.
func Load(conn *sql.DB) (Snapshot, error) {
	var s Snapshot
	if err := conn.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&s.Total); err != nil {
		return s, err
	}
	rows, err := conn.Query(`SELECT * FROM records ORDER BY rowid LIMIT ?`, previewLimit)
	if err != nil {
		return s, err
	}
	defer rows.Close()

	if s.Columns, err = rows.Columns(); err != nil {
		return s, err
	}
	for rows.Next() {
		vals := make([]any, len(s.Columns))
		ptrs := make([]any, len(vals))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return s, err
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			switch t := v.(type) {
			case nil:
				row[i] = "NULL"
			case []byte:
				row[i] = string(t)
			default:
				row[i] = fmt.Sprint(t)
			}
		}
		s.Rows = append(s.Rows, row)
	}
	return s, rows.Err()
}
