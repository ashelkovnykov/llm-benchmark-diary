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

const (
	// Dir is the directory that holds the database files.
	Dir = "db"
	// TableName is the table exposed by the schema.
	TableName = "records"

	fileExt = ".db"
)

//go:embed schema.sql
var schemaFS embed.FS

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// DB wraps an open SQLite database file.
type DB struct {
	Name string
	conn *sql.DB
}

// Table holds display-ready rows from the records table.
type Table struct {
	Columns []string
	Rows    [][]string
}

// List returns the names of the databases stored in Dir.
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
		if name := e.Name(); strings.HasSuffix(name, fileExt) {
			names = append(names, strings.TrimSuffix(name, fileExt))
		}
	}
	sort.Strings(names)
	return names, nil
}

// ValidateNewName reports whether name may be used for a new database.
func ValidateNewName(name string) error {
	switch {
	case name == "":
		return errors.New("name cannot be empty")
	case !namePattern.MatchString(name):
		return errors.New("name must contain only letters and numbers")
	case Exists(name):
		return fmt.Errorf("a database named %q already exists", name)
	}
	return nil
}

// Exists reports whether a database with the given name is present in Dir.
func Exists(name string) bool {
	_, err := os.Stat(filepath.Join(Dir, name+fileExt))
	return err == nil
}

// Open opens an existing database from Dir.
func Open(name string) (*DB, error) {
	conn, err := sql.Open("sqlite", filepath.Join(Dir, name+fileExt))
	if err != nil {
		return nil, err
	}
	d := &DB{Name: name, conn: conn}
	if err := conn.Ping(); err != nil {
		d.Close()
		return nil, fmt.Errorf("open %q: %w", name, err)
	}
	return d, nil
}

// Create builds a new database in Dir from the embedded schema and seeds it
// with dummy records. The returned database is open and ready to use.
func Create(name string) (d *DB, err error) {
	if err := ValidateNewName(name); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return nil, err
	}

	path := filepath.Join(Dir, name+fileExt)
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	d = &DB{Name: name, conn: conn}
	defer func() {
		if err != nil {
			d.Close()
			os.Remove(path)
		}
	}()

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return nil, err
	}
	if _, err = conn.Exec(string(schema)); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err = seedRecords(conn, dummyCount); err != nil {
		return nil, err
	}
	return d, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	if d == nil || d.conn == nil {
		return nil
	}
	return d.conn.Close()
}

// TotalRecords returns the number of rows in the records table.
func (d *DB) TotalRecords() (int, error) {
	var n int
	err := d.conn.QueryRow("SELECT COUNT(*) FROM " + quoteIdent(TableName)).Scan(&n)
	return n, err
}

// Records returns up to limit rows from the records table.
func (d *DB) Records(limit int) (Table, error) {
	rows, err := d.conn.Query(fmt.Sprintf("SELECT * FROM %s LIMIT %d", quoteIdent(TableName), limit))
	if err != nil {
		return Table{}, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return Table{}, err
	}
	data := Table{Columns: cols}
	for rows.Next() {
		raw := make([]any, len(cols))
		dest := make([]any, len(cols))
		for i := range raw {
			dest[i] = &raw[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return Table{}, err
		}
		line := make([]string, len(cols))
		for i, v := range raw {
			line[i] = formatValue(v)
		}
		data.Rows = append(data.Rows, line)
	}
	return data, rows.Err()
}

func formatValue(v any) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
