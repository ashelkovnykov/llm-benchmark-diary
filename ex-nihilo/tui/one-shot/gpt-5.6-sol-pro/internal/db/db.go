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

// Dir is where generated databases live, relative to the process CWD.
const Dir = "./db"

const ext = ".db"

//go:embed schema.sql
var schema string

var nameRe = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// List returns the names (without extension) of the databases in Dir,
// sorted alphabetically. Dir is created if it does not exist.
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
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ext) {
			names = append(names, strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))
		}
	}
	sort.Strings(names)
	return names, nil
}

// ValidateName checks a candidate database name against the existing set.
func ValidateName(name string, existing []string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name cannot be empty")
	}
	if !nameRe.MatchString(name) {
		return errors.New("name must contain only alphanumeric characters")
	}
	for _, e := range existing {
		if strings.EqualFold(e, name) {
			return fmt.Errorf("a database named %q already exists", name)
		}
	}
	return nil
}

func path(name string) string {
	return filepath.Join(Dir, name+ext)
}

// Open opens an existing database.
func Open(name string) (*sql.DB, error) {
	return sql.Open("sqlite", path(name))
}

// Create creates a new database, applies the schema, seeds dummy records,
// and returns the open handle.
func Create(name string) (*sql.DB, error) {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return nil, err
	}
	conn, err := sql.Open("sqlite", path(name))
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("applying schema: %w", err)
	}
	if err := Seed(conn, 15); err != nil {
		conn.Close()
		return nil, fmt.Errorf("seeding records: %w", err)
	}
	return conn, nil
}

// Seed inserts n dummy records. Placeholder columns only -- this will need to
// be reworked once the real schema is known.
func Seed(conn *sql.DB, n int) error {
	for i := 1; i <= n; i++ {
		_, err := conn.Exec(
			`INSERT INTO records (name, value) VALUES (?, ?)`,
			fmt.Sprintf("record_%d", i),
			fmt.Sprintf("value_%d", i),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// Snapshot is a read-only view of the records table for display purposes.
type Snapshot struct {
	Count   int
	Columns []string
	Rows    [][]string
}

// SnapshotOf reads the total record count plus up to limit rows.
func SnapshotOf(conn *sql.DB, limit int) (*Snapshot, error) {
	s := &Snapshot{}

	if err := conn.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&s.Count); err != nil {
		return nil, err
	}

	rows, err := conn.Query(fmt.Sprintf(`SELECT * FROM records LIMIT %d`, limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	s.Columns = cols

	for rows.Next() {
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		out := make([]string, len(cols))
		for i, v := range raw {
			out[i] = format(v)
		}
		s.Rows = append(s.Rows, out)
	}

	return s, rows.Err()
}

func format(v any) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}
