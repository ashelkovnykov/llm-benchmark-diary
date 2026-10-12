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
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

const (
	// DirName is the directory database files live in, relative to the cwd.
	DirName = "db"
	// Ext is the file extension used for database files.
	Ext = ".db"
	// DummyRows is how many placeholder records a new database is seeded with.
	DummyRows = 15
)

var validName = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// Info describes a database file available on disk.
type Info struct {
	Name string
	Path string
}

// Dir returns the database directory.
func Dir() string { return DirName }

// EnsureDir creates the database directory if it does not exist.
func EnsureDir() error {
	return os.MkdirAll(DirName, 0o755)
}

// Path returns the on-disk path for a database name.
func Path(name string) string {
	return filepath.Join(DirName, name+Ext)
}

// List returns the databases available in the database directory.
func List() ([]Info, error) {
	entries, err := os.ReadDir(DirName)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	infos := make([]Info, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if !strings.EqualFold(ext, Ext) {
			continue
		}
		infos = append(infos, Info{
			Name: strings.TrimSuffix(e.Name(), ext),
			Path: filepath.Join(DirName, e.Name()),
		})
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos, nil
}

// ValidateName reports why name cannot be used for a new database, or nil.
func ValidateName(name string, existing []Info) error {
	name = strings.TrimSpace(name)

	switch {
	case name == "":
		return errors.New("name cannot be empty")
	case !validName.MatchString(name):
		return errors.New("name may only contain letters and numbers")
	}

	for _, info := range existing {
		if strings.EqualFold(info.Name, name) {
			return fmt.Errorf("a database named %q already exists", name)
		}
	}
	if _, err := os.Stat(Path(name)); err == nil {
		return fmt.Errorf("a database named %q already exists", name)
	}

	return nil
}

// Open opens an existing database by name.
func Open(name string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", Path(name))

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// Create builds a new database from schema.sql and seeds it with dummy records.
func Create(name string) (*sql.DB, error) {
	if err := EnsureDir(); err != nil {
		return nil, err
	}

	path := Path(name)
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("a database named %q already exists", name)
	}

	conn, err := Open(name)
	if err != nil {
		return nil, err
	}

	if err := applySchema(conn); err != nil {
		conn.Close()
		os.Remove(path)
		return nil, err
	}
	if err := seedDummyRecords(conn, DummyRows); err != nil {
		conn.Close()
		os.Remove(path)
		return nil, err
	}

	return conn, nil
}

func applySchema(conn *sql.DB) error {
	stmt, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	if _, err := conn.Exec(string(stmt)); err != nil {
		return fmt.Errorf("applying schema: %w", err)
	}
	return nil
}

// seedDummyRecords inserts n placeholder rows. Delete once real import exists.
func seedDummyRecords(conn *sql.DB, n int) error {
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
		if _, err := stmt.Exec(fmt.Sprintf("record-%02d", i), i); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Summary is a snapshot of a database used for the preview screen.
type Summary struct {
	Columns []string
	Rows    [][]string
	Total   int
}

// Preview reads the record count and up to limit rows from the records table.
func Preview(conn *sql.DB, limit int) (*Summary, error) {
	s := &Summary{}

	if err := conn.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&s.Total); err != nil {
		return nil, fmt.Errorf("counting records: %w", err)
	}

	rows, err := conn.Query(`SELECT * FROM records ORDER BY rowid LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("reading records: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	s.Columns = cols

	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		row := make([]string, len(cols))
		for i, v := range values {
			row[i] = formatValue(v)
		}
		s.Rows = append(s.Rows, row)
	}

	return s, rows.Err()
}

func formatValue(v any) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}
