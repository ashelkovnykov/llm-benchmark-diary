// Package db provides access to the SQLite databases stored under ./db.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// Dir is the directory that holds the SQLite databases.
const Dir = "db"

// ext is the file extension used for database files.
const ext = ".sqlite"

// seedRows is the number of dummy rows inserted into a freshly created DB.
const seedRows = 15

var validName = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// Store is an open database handle.
type Store struct {
	Name string
	Path string
	db   *sql.DB
}

// Close releases the underlying database handle.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// List returns the names of the databases available in ./db, sorted.
// The directory is created if it does not yet exist.
func List() ([]string, error) {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(Dir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ext) {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ext))
	}
	sort.Strings(names)
	return names, nil
}

// ValidateName reports why name is unusable as a database name, or nil if it
// is acceptable. existing is the current list of database names.
func ValidateName(name string, existing []string) error {
	switch {
	case strings.TrimSpace(name) == "":
		return fmt.Errorf("name cannot be empty")
	case !validName.MatchString(name):
		return fmt.Errorf("name may contain only letters and digits")
	}

	for _, e := range existing {
		if e == name {
			return fmt.Errorf("a database named %q already exists", name)
		}
	}
	return nil
}

// Open opens an existing database.
func Open(name string) (*Store, error) {
	return open(filepath.Join(Dir, name+ext))
}

// Create builds a new database, applies the schema and seeds dummy rows.
func Create(name string) (*Store, error) {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return nil, err
	}

	path := filepath.Join(Dir, name+ext)
	store, err := open(path)
	if err != nil {
		return nil, err
	}

	if err := store.applySchema(); err != nil {
		store.Close()
		os.Remove(path)
		return nil, err
	}
	if err := store.seed(seedRows); err != nil {
		store.Close()
		os.Remove(path)
		return nil, err
	}
	return store, nil
}

func open(path string) (*Store, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return &Store{
		Name: strings.TrimSuffix(filepath.Base(path), ext),
		Path: path,
		db:   sqlDB,
	}, nil
}

func (s *Store) applySchema() error {
	raw, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	_, err = s.db.Exec(string(raw))
	return err
}

func (s *Store) seed(n int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO records (rowid, title, body) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= n; i++ {
		if _, err := stmt.Exec(
			i,
			fmt.Sprintf("Dummy record %d", i),
			fmt.Sprintf("Placeholder body for record %d.", i),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Count returns the number of rows in the records table.
func (s *Store) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&n)
	return n, err
}

// Rows returns up to limit rows as strings, together with the column headers.
func (s *Store) Rows(limit int) ([]string, [][]string, error) {
	rows, err := s.db.Query(fmt.Sprintf(`SELECT * FROM records LIMIT %d`, limit))
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
			rec[i] = format(v)
		}
		out = append(out, rec)
	}
	return cols, out, rows.Err()
}

func format(v any) string {
	switch x := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(x)
	default:
		return fmt.Sprint(x)
	}
}
