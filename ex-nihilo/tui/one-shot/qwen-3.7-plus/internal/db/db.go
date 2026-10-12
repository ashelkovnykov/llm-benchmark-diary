package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

const dbDir = "./db"

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func EnsureDir() error {
	return os.MkdirAll(dbDir, 0755)
}

func ListDBs() ([]string, error) {
	if err := EnsureDir(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dbDir)
	if err != nil {
		return nil, err
	}
	var dbs []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".db" {
			dbs = append(dbs, e.Name())
		}
	}
	return dbs, nil
}

func ValidateName(name string, existing []string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("name must be alphanumeric")
	}
	for _, e := range existing {
		if e == name+".db" {
			return fmt.Errorf("name already exists")
		}
	}
	return nil
}

func CreateDB(name string) (*sql.DB, error) {
	path := filepath.Join(dbDir, name+".db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(schemaSQL); err != nil {
		conn.Close()
		return nil, err
	}

	stmt, err := conn.Prepare("INSERT INTO records (value) VALUES (?)")
	if err != nil {
		conn.Close()
		return nil, err
	}
	defer stmt.Close()

	for i := 1; i <= 15; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("dummy_data_%d", i)); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return conn, nil
}

func OpenDB(name string) (*sql.DB, error) {
	path := filepath.Join(dbDir, name)
	return sql.Open("sqlite", path)
}

type Record struct {
	RowID int
	Value string
}

func GetStats(conn *sql.DB) (int, []Record, error) {
	var count int
	if err := conn.QueryRow("SELECT COUNT(*) FROM records").Scan(&count); err != nil {
		return 0, nil, err
	}

	rows, err := conn.Query("SELECT rowid, value FROM records LIMIT 10")
	if err != nil {
		return count, nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Value); err != nil {
			return count, nil, err
		}
		records = append(records, r)
	}
	return count, records, nil
}
