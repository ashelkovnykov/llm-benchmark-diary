package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

const DBDir = "./db"

type Record struct {
	RowID     int64
	Title     string
	Value     float64
	CreatedAt string
}

func EnsureDBDir() error {
	return os.MkdirAll(DBDir, 0755)
}

func ListDatabases() ([]string, error) {
	if err := EnsureDBDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(DBDir)
	if err != nil {
		return nil, err
	}

	var dbs []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".db") {
			dbs = append(dbs, entry.Name())
		}
	}
	return dbs, nil
}

func ValidateDBName(name string, existing []string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", name)
	if !matched {
		return fmt.Errorf("name must contain only alphanumeric characters")
	}

	target := name + ".db"
	for _, db := range existing {
		if strings.EqualFold(db, target) {
			return fmt.Errorf("database '%s' already exists", target)
		}
	}

	return nil
}

func CreateDatabase(name string) (string, error) {
	filename := name + ".db"
	fullPath := filepath.Join(DBDir, filename)

	conn, err := sql.Open("sqlite3", fullPath)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if _, err := conn.Exec(schemaSQL); err != nil {
		return "", fmt.Errorf("failed executing schema: %w", err)
	}

	tx, err := conn.Begin()
	if err != nil {
		return "", err
	}

	stmt, err := tx.Prepare("INSERT INTO records (title, value) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return "", err
	}
	defer stmt.Close()

	for i := 1; i <= 15; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("Dummy Record %02d", i), float64(i)*10.5); err != nil {
			tx.Rollback()
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return filename, nil
}

func LoadData(filename string) (int, []Record, error) {
	fullPath := filepath.Join(DBDir, filename)
	conn, err := sql.Open("sqlite3", fullPath)
	if err != nil {
		return 0, nil, err
	}
	defer conn.Close()

	var total int
	err = conn.QueryRow("SELECT COUNT(*) FROM records").Scan(&total)
	if err != nil {
		return 0, nil, fmt.Errorf("failed counting records: %w", err)
	}

	rows, err := conn.Query("SELECT rowid, title, value, created_at FROM records ORDER BY rowid ASC LIMIT 10")
	if err != nil {
		return 0, nil, fmt.Errorf("failed querying rows: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Title, &r.Value, &r.CreatedAt); err != nil {
			return 0, nil, err
		}
		records = append(records, r)
	}

	return total, records, nil
}
