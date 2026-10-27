package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

var alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

type Record struct {
	RowID int
	Name  string
	Value string
}

type DBData struct {
	Name       string
	TotalCount int
	Records    []Record
}

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func ListDBs(dir string) ([]string, error) {
	if err := EnsureDir(dir); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dbs []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".db") {
			dbs = append(dbs, strings.TrimSuffix(entry.Name(), ".db"))
		}
	}
	return dbs, nil
}

func ValidateDBName(dir, name string) error {
	if name == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	if !alphanumericRegex.MatchString(name) {
		return fmt.Errorf("database name must contain only alphanumeric characters")
	}

	existing, err := ListDBs(dir)
	if err != nil {
		return err
	}
	for _, dbName := range existing {
		if strings.EqualFold(dbName, name) {
			return fmt.Errorf("database '%s' already exists", name)
		}
	}
	return nil
}

func CreateDB(dir, name string) error {
	if err := ValidateDBName(dir, name); err != nil {
		return err
	}

	dbPath := filepath.Join(dir, name+".db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Exec(schemaSQL); err != nil {
		return err
	}

	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO records (name, value) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i := 1; i <= 15; i++ {
		_, err := stmt.Exec(fmt.Sprintf("Item_%d", i), fmt.Sprintf("Dummy_Value_%d", i*10))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func LoadDBData(dir, name string) (*DBData, error) {
	dbPath := filepath.Join(dir, name+".db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var total int
	if err := conn.QueryRow("SELECT COUNT(*) FROM records").Scan(&total); err != nil {
		return nil, err
	}

	rows, err := conn.Query("SELECT rowid, name, value FROM records ORDER BY rowid ASC LIMIT 10")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Name, &r.Value); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return &DBData{
		Name:       name,
		TotalCount: total,
		Records:    records,
	}, rows.Err()
}
