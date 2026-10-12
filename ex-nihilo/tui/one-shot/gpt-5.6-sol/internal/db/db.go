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

const (
	Dir        = "db"
	Ext        = ".db"
	RecordsTbl = "records"
)

var validName = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// EnsureDir creates the database directory if it does not already exist.
func EnsureDir() error {
	return os.MkdirAll(Dir, 0o755)
}

// path returns the on-disk path for a named database.
func path(name string) string {
	return filepath.Join(Dir, name+Ext)
}

// List returns the names of every database currently in the db directory.
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
		if e.IsDir() || !strings.HasSuffix(e.Name(), Ext) {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), Ext))
	}
	sort.Strings(names)
	return names, nil
}

// Exists reports whether a database with the given name exists.
func Exists(name string) bool {
	_, err := os.Stat(path(name))
	return err == nil
}

// ValidateName checks a candidate database name and returns a human readable
// error describing the first problem found.
func ValidateName(name string) error {
	switch {
	case strings.TrimSpace(name) == "":
		return fmt.Errorf("name cannot be empty")
	case !validName.MatchString(name):
		return fmt.Errorf("name may only contain letters and numbers")
	case Exists(name):
		return fmt.Errorf("a database named %q already exists", name)
	}
	return nil
}

// Create builds a new database from the embedded schema and seeds it with
// dummy records.
func Create(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	if err := EnsureDir(); err != nil {
		return err
	}

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	conn, err := sql.Open("sqlite", path(name))
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Exec(string(schema)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if err := seed(conn, 15); err != nil {
		return fmt.Errorf("seed records: %w", err)
	}
	return nil
}

// Open connects to an existing database.
func Open(name string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path(name))
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// CountRecords returns the number of rows in the records table.
func CountRecords(conn *sql.DB) (int, error) {
	var n int
	err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", RecordsTbl)).Scan(&n)
	return n, err
}

// Preview returns the column names and up to limit rows from the records table.
func Preview(conn *sql.DB, limit int) ([]string, [][]string, error) {
	rows, err := conn.Query(fmt.Sprintf("SELECT * FROM %s LIMIT %d", RecordsTbl, limit))
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

// seed inserts n dummy records. Tied to the placeholder schema.
func seed(conn *sql.DB, n int) error {
	names := []string{
		"alpha", "bravo", "charlie", "delta", "echo",
		"foxtrot", "golf", "hotel", "india", "juliet",
		"kilo", "lima", "mike", "november", "oscar",
	}
	cats := []string{"red", "green", "blue", "yellow", "purple"}

	for i := 0; i < n; i++ {
		name := fmt.Sprintf("%s-%d", names[i%len(names)], i+1)
		cat := cats[i%len(cats)]
		if _, err := conn.Exec(
			"INSERT INTO records (name, category, value) VALUES (?, ?, ?)",
			name, cat, (i+1)*10,
		); err != nil {
			return err
		}
	}
	return nil
}

func format(v any) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
