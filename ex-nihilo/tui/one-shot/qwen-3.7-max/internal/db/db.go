package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

const dbDir = "./db"

func EnsureDBDir() error {
	return os.MkdirAll(dbDir, 0755)
}

func ListDBs() ([]string, error) {
	entries, err := os.ReadDir(dbDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var dbs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			dbs = append(dbs, entry.Name()[:len(entry.Name())-3])
		}
	}
	return dbs, nil
}

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func ValidateName(name string, existing []string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("name must contain only alphanumeric characters")
	}
	for _, ex := range existing {
		if ex == name {
			return fmt.Errorf("database '%s' already exists", name)
		}
	}
	return nil
}

func CreateDB(name string) (*sql.DB, error) {
	if err := EnsureDBDir(); err != nil {
		return nil, err
	}

	path := filepath.Join(dbDir, name+".db")
	os.Remove(path)

	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	schemaBytes, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return nil, err
	}

	_, err = database.Exec(string(schemaBytes))
	if err != nil {
		return nil, err
	}

	tx, err := database.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("INSERT INTO records (name, value, created_at) VALUES (?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i := 1; i <= 15; i++ {
		_, err = stmt.Exec(fmt.Sprintf("Dummy Record %d", i), float64(i)*1.5, time.Now().Format(time.RFC3339))
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return database, tx.Commit()
}

func OpenDB(name string) (*sql.DB, error) {
	path := filepath.Join(dbDir, name+".db")
	return sql.Open("sqlite", path)
}

type Record struct {
	RowID     int
	Name      string
	Value     float64
	CreatedAt string
}

func GetStats(database *sql.DB) (int, error) {
	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM records").Scan(&count)
	return count, err
}

func GetRecords(database *sql.DB, limit int) ([]Record, error) {
	rows, err := database.Query("SELECT rowid, name, value, created_at FROM records LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Name, &r.Value, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
