package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

const dbDir = "db"

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// ListDatabases returns the names (without extension) of all SQLite databases
// found in the project's ./db/ directory. The directory is created if missing.
func ListDatabases() ([]string, error) {
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	entries, err := os.ReadDir(dbDir)
	if err != nil {
		return nil, fmt.Errorf("read db dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".db") {
			names = append(names, strings.TrimSuffix(e.Name(), ".db"))
		}
	}
	sort.Strings(names)
	return names, nil
}

// ValidateName checks that the candidate database name is non-empty,
// alphanumeric, and does not collide with an existing database.
func ValidateName(name string, existing []string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("name must contain only alphanumeric characters (no spaces or symbols)")
	}
	for _, e := range existing {
		if e == name {
			return fmt.Errorf("a database named %q already exists", e)
		}
	}
	return nil
}

// CreateDatabase creates a new SQLite database at ./db/<name>.db, applies the
// embedded schema, and populates the records table with 15 dummy rows.
func CreateDatabase(name string) (string, error) {
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return "", fmt.Errorf("create db dir: %w", err)
	}
	path := filepath.Join(dbDir, name+".db")
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("a database named %q already exists", name)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return "", fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Exec(schemaSQL); err != nil {
		return "", fmt.Errorf("apply schema: %w", err)
	}

	tx, err := conn.Begin()
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	stmt, err := tx.Prepare("INSERT INTO records (title, description, amount) VALUES (?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		return "", fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for i := 1; i <= 15; i++ {
		title := fmt.Sprintf("Record %02d", i)
		description := fmt.Sprintf("This is dummy record number %d", i)
		amount := float64(i) * 10.5
		if _, err := stmt.Exec(title, description, amount); err != nil {
			_ = tx.Rollback()
			return "", fmt.Errorf("insert record %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}
	return path, nil
}

// OpenDatabase opens a connection to an existing database by base name.
func OpenDatabase(name string) (*sql.DB, error) {
	path := filepath.Join(dbDir, name+".db")
	return sql.Open("sqlite", path)
}

// Record is a minimal representation of a row in the records table.
type Record struct {
	RowID       int64
	Title       string
	Description string
	Amount      float64
	CreatedAt   string
}

// CountRecords returns the total number of rows in the records table.
func CountRecords(conn *sql.DB) (int, error) {
	var count int
	err := conn.QueryRow("SELECT COUNT(*) FROM records").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count records: %w", err)
	}
	return count, nil
}

// GetRecords returns up to `limit` rows from the records table, ordered by rowid.
func GetRecords(conn *sql.DB, limit int) ([]Record, error) {
	rows, err := conn.Query(`SELECT rowid, title, description, amount, created_at FROM records ORDER BY rowid ASC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query records: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Title, &r.Description, &r.Amount, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan record: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate records: %w", err)
	}
	return records, nil
}
