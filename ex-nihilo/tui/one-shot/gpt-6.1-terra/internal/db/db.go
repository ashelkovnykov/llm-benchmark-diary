package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	directory  = "./db"
	schemaPath = "./internal/db/schema.sql"
)

var validName = regexp.MustCompile(`^[A-Za-z0-9]+$`)

type Snapshot struct {
	Name    string
	Count   int64
	Columns []string
	Rows    [][]string
}

func List() ([]string, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read database directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".db") {
			names = append(names, strings.TrimSuffix(entry.Name(), ".db"))
		}
	}

	sort.Strings(names)
	return names, nil
}

func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("name must contain only ASCII letters and digits")
	}

	names, err := List()
	if err != nil {
		return err
	}
	for _, existing := range names {
		if existing == name {
			return fmt.Errorf("a database with that name already exists")
		}
	}
	return nil
}

func Create(name string) (err error) {
	if err := ValidateName(name); err != nil {
		return err
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	path := filepath.Join(directory, name+".db")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("a database with that name already exists")
		}
		return fmt.Errorf("create database file: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return fmt.Errorf("close database file: %w", err)
	}

	// Remove an incomplete database if schema initialization fails.
	defer func() {
		if err != nil {
			os.Remove(path)
		}
	}()

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("begin initialization: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(schema)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	for i := 1; i <= 15; i++ {
		if _, err := tx.Exec("INSERT INTO records (rowid) VALUES (?)", i); err != nil {
			return fmt.Errorf("insert dummy record %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit initialization: %w", err)
	}
	return nil
}

func Load(name string) (Snapshot, error) {
	result := Snapshot{Name: name}

	// Only load files returned by List, never an arbitrary path.
	names, err := List()
	if err != nil {
		return result, err
	}
	found := false
	for _, candidate := range names {
		if candidate == name {
			found = true
			break
		}
	}
	if !found {
		return result, fmt.Errorf("database %q does not exist", name)
	}

	path := filepath.Join(directory, name+".db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return result, fmt.Errorf("open database: %w", err)
	}
	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		return result, fmt.Errorf("begin read: %w", err)
	}
	defer tx.Rollback()

	if err := tx.QueryRow("SELECT COUNT(*) FROM records").Scan(&result.Count); err != nil {
		return result, fmt.Errorf("count records: %w", err)
	}

	rows, err := tx.Query("SELECT * FROM records ORDER BY rowid LIMIT 10")
	if err != nil {
		return result, fmt.Errorf("query records: %w", err)
	}

	result.Columns, err = rows.Columns()
	if err != nil {
		rows.Close()
		return result, fmt.Errorf("read columns: %w", err)
	}

	for rows.Next() {
		values := make([]any, len(result.Columns))
		destinations := make([]any, len(values))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := rows.Scan(destinations...); err != nil {
			rows.Close()
			return result, fmt.Errorf("scan row: %w", err)
		}

		record := make([]string, len(values))
		for i, value := range values {
			switch v := value.(type) {
			case nil:
				record[i] = "NULL"
			case []byte:
				record[i] = string(v)
			default:
				record[i] = fmt.Sprint(v)
			}
		}
		result.Rows = append(result.Rows, record)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, fmt.Errorf("read rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("finish read: %w", err)
	}
	return result, nil
}
