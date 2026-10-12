package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

const DBDir = "db"

// ListDBs returns the filenames (without extension) of all .db files in ./db/.
func ListDBs() ([]string, error) {
	entries, err := os.ReadDir(DBDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
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
	return names, nil
}

// DBPath returns the full path for a given database name.
func DBPath(name string) string {
	return filepath.Join(DBDir, name+".db")
}

// CreateDB creates a new SQLite database with the embedded schema and seeds it.
func CreateDB(name string) error {
	path := DBPath(name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("database %q already exists", name)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("reading schema: %w", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("applying schema: %w", err)
	}

	if err := Seed(db, 15); err != nil {
		return fmt.Errorf("seeding: %w", err)
	}

	return nil
}

// Open opens an existing database by name.
func Open(name string) (*sql.DB, error) {
	path := DBPath(name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("database %q not found", name)
	}
	return sql.Open("sqlite", path)
}

// Record represents a single row from the records table.
type Record struct {
	RowID     int
	Name      string
	Value     string
	CreatedAt string
}

// FetchRecords returns the total count and up to `limit` rows.
func FetchRecords(db *sql.DB, limit int) (int, []Record, error) {
	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM records").Scan(&total); err != nil {
		return 0, nil, err
	}

	rows, err := db.Query("SELECT rowid, name, value, created_at FROM records ORDER BY rowid LIMIT ?", limit)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Name, &r.Value, &r.CreatedAt); err != nil {
			return 0, nil, err
		}
		records = append(records, r)
	}
	return total, records, rows.Err()
}
