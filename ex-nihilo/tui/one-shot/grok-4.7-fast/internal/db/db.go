package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	path string
}

func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return &DB{DB: sqlDB, path: path}, nil
}

func (d *DB) Close() error {
	return d.DB.Close()
}

func (d *DB) Path() string {
	return d.path
}

func ListDatabases(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var dbs []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".db") {
			dbs = append(dbs, strings.TrimSuffix(e.Name(), ".db"))
		}
	}
	return dbs, nil
}

var nameRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func ValidateName(name string, existing []string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !nameRe.MatchString(name) {
		return fmt.Errorf("name must contain only alphanumeric characters")
	}
	for _, e := range existing {
		if e == name {
			return fmt.Errorf("database '%s' already exists", name)
		}
	}
	return nil
}

func CreateDatabase(dir, name string, schemaPath string) (*DB, error) {
	path := filepath.Join(dir, name+".db")
	db, err := Open(path)
	if err != nil {
		return nil, err
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("read schema: %w", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, fmt.Errorf("exec schema: %w", err)
	}
	// Populate with 15 dummy records
	for i := 1; i <= 15; i++ {
		_, err := db.Exec("INSERT INTO records (name, value, created_at) VALUES (?, ?, datetime('now'))",
			fmt.Sprintf("Record %d", i), i*10)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("insert dummy: %w", err)
		}
	}
	return db, nil
}

func (d *DB) RecordCount() (int, error) {
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM records").Scan(&count)
	return count, err
}

func (d *DB) FirstNRecords(n int) ([][]string, []string, error) {
	rows, err := d.Query("SELECT * FROM records LIMIT ?", n)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var result [][]string
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			if v == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprint(v)
			}
		}
		result = append(result, row)
	}
	return result, cols, rows.Err()
}
