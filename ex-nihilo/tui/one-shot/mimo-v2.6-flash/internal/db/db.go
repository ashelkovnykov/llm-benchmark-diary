// Package db lists, creates, opens and queries the project's SQLite databases.
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
	// Dir is the directory, relative to the working directory, holding the databases.
	Dir = "db"
	// Extension is appended to every database file name.
	Extension = ".db"
	// SeedCount is the number of dummy records inserted into a new database.
	SeedCount = 15
	// PreviewLimit is the default number of rows returned by Preview.
	PreviewLimit = 10
	// driver is the registered SQLite driver name.
	driver = "sqlite"
)

var namePattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// DirPath returns the path of the database directory.
func DirPath() string { return filepath.Join(".", Dir) }

// Path returns the file path of the database with the given name.
func Path(name string) string { return filepath.Join(DirPath(), name+Extension) }

// List returns the sorted names of every database in the database directory.
// A missing directory is not an error; it yields an empty list.
func List() ([]string, error) {
	entries, err := os.ReadDir(DirPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", DirPath(), err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if n := e.Name(); strings.HasSuffix(n, Extension) {
			names = append(names, strings.TrimSuffix(n, Extension))
		}
	}
	sort.Strings(names)
	return names, nil
}

// Exists reports whether a database with the given name already exists.
func Exists(name string) bool {
	_, err := os.Stat(Path(name))
	return err == nil
}

// ValidateName reports why name is not a usable database name, or nil.
func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("the name cannot be empty")
	}
	if !namePattern.MatchString(name) {
		return errors.New("the name must contain only alphanumeric characters (a-z, A-Z, 0-9)")
	}
	if Exists(name) {
		return fmt.Errorf("a database named %q already exists", name)
	}
	return nil
}

// Create validates name, builds the database from schema.sql and seeds it.
// On failure the partially written file is removed.
func Create(name string) (string, error) {
	name = strings.TrimSpace(name)
	if err := ValidateName(name); err != nil {
		return "", fmt.Errorf("invalid name: %w", err)
	}
	if err := os.MkdirAll(DirPath(), 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", DirPath(), err)
	}
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return "", fmt.Errorf("read schema: %w", err)
	}
	path := Path(name)
	conn, err := open(path)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", name+Extension, err)
	}
	if _, err := conn.Exec(string(schema)); err != nil {
		conn.Close()
		os.Remove(path)
		return "", fmt.Errorf("apply schema: %w", err)
	}
	if err := seed(conn); err != nil {
		conn.Close()
		os.Remove(path)
		return "", fmt.Errorf("seed records: %w", err)
	}
	if err := conn.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}

// Open opens an existing database and verifies it is reachable.
func Open(name string) (*sql.DB, error) {
	if !Exists(name) {
		return nil, fmt.Errorf("database %q does not exist", name)
	}
	conn, err := open(Path(name))
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", name, err)
	}
	return conn, nil
}

func open(path string) (*sql.DB, error) {
	conn, err := sql.Open(driver, path)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// seed inserts SeedCount dummy records into the records table.
func seed(conn *sql.DB) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO records (name, category, value, notes) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	categories := []string{"alpha", "beta", "gamma", "delta"}
	for i := 1; i <= SeedCount; i++ {
		if _, err := stmt.Exec(
			fmt.Sprintf("record-%02d", i),
			categories[i%len(categories)],
			i*10,
			fmt.Sprintf("dummy record %d", i),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}
