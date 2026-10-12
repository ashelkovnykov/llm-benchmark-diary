package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

const (
	// Dir is the directory where databases are stored, relative to the CWD.
	Dir = "db"
	// Ext is the file extension used for database files.
	Ext = ".db"
	// SeedRows is the number of dummy records inserted into a new database.
	SeedRows = 15
)

var nameRe = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// Path returns the on-disk path for a database name.
func Path(name string) string {
	return filepath.Join(Dir, name+Ext)
}

// List returns the names (without extension) of every database in Dir, sorted.
// A missing Dir is not an error; it yields an empty list.
func List() ([]string, error) {
	entries, err := os.ReadDir(Dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), Ext) {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), Ext))
	}
	sort.Strings(names)
	return names, nil
}

// ValidateName reports why name is not usable for a new database.
// It returns nil if the name is acceptable.
func ValidateName(name string, existing []string) error {
	switch {
	case name == "":
		return errors.New("name cannot be empty")
	case !nameRe.MatchString(name):
		return errors.New("name must contain only alphanumeric characters")
	}
	for _, e := range existing {
		if strings.EqualFold(e, name) {
			return fmt.Errorf("a database named %q already exists", name)
		}
	}
	return nil
}

// Create builds a new database file, applies the embedded schema, and seeds it
// with SeedRows dummy records. On any failure the partial file is removed.
func Create(name string) error {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return err
	}

	path := Path(name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("a database named %q already exists", name)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return err
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer conn.Close()

	cleanup := func(err error) error {
		conn.Close()
		os.Remove(path)
		return err
	}

	if _, err := conn.Exec(string(schema)); err != nil {
		return cleanup(fmt.Errorf("applying schema: %w", err))
	}
	if err := seed(conn, SeedRows); err != nil {
		return cleanup(fmt.Errorf("seeding: %w", err))
	}
	return nil
}

// Open opens an existing database by name.
func Open(name string) (*sql.DB, error) {
	path := Path(name)
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("cannot open %q: %w", name, err)
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// Table is a snapshot of the records table.
type Table struct {
	Columns []string
	Rows    [][]string
	Total   int
}

// Load reads the total record count and up to limit rows from the records table.
func Load(conn *sql.DB, limit int) (*Table, error) {
	t := &Table{}

	if err := conn.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&t.Total); err != nil {
		return nil, fmt.Errorf("counting records: %w", err)
	}

	rows, err := conn.Query(
		`SELECT * FROM records ORDER BY rowid LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("reading records: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	t.Columns = cols

	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		rec := make([]string, len(cols))
		for i, v := range vals {
			rec[i] = format(v)
		}
		t.Rows = append(t.Rows, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return t, nil
}

func seed(conn *sql.DB, n int) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO records (name, value) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= n; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("record %d", i), i*10); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func format(v any) string {
	switch x := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(x)
	case string:
		return x
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	default:
		return fmt.Sprintf("%v", x)
	}
}
