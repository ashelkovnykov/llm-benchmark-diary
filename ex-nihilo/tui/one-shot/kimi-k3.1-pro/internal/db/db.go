package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

const directory = "./db"

var validName = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

//go:embed schema.sql
var schema string

type DB struct {
	sql *sql.DB
}

type Snapshot struct {
	Total   int64
	Columns []string
	Rows    [][]string
}

func List() ([]string, error) {
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".db") {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
	}

	sort.Strings(names)
	return names, nil
}

func Create(name string) error {
	if name == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("database name may contain only alphanumeric characters")
	}

	names, err := List()
	if err != nil {
		return err
	}
	for _, existing := range names {
		if existing == name {
			return fmt.Errorf("a database named %q already exists", name)
		}
	}

	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}

	database, err := sql.Open("sqlite", filepath.Join(directory, name+".db"))
	if err != nil {
		return err
	}
	defer database.Close()

	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(schema); err != nil {
		return err
	}

	for i := 1; i <= 15; i++ {
		if _, err := tx.Exec(
			"INSERT INTO records (name, value) VALUES (?, ?)",
			fmt.Sprintf("Dummy record %d", i),
			i,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func Open(name string) (*DB, error) {
	database, err := sql.Open("sqlite", filepath.Join(directory, name+".db"))
	if err != nil {
		return nil, err
	}

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, err
	}

	return &DB{sql: database}, nil
}

func (d *DB) Close() error {
	if d == nil || d.sql == nil {
		return nil
	}
	return d.sql.Close()
}

func (d *DB) Snapshot() (Snapshot, error) {
	var snapshot Snapshot

	if err := d.sql.QueryRow("SELECT COUNT(*) FROM records").Scan(&snapshot.Total); err != nil {
		return snapshot, err
	}

	rows, err := d.sql.Query("SELECT * FROM records ORDER BY rowid LIMIT 10")
	if err != nil {
		return snapshot, err
	}
	defer rows.Close()

	snapshot.Columns, err = rows.Columns()
	if err != nil {
		return snapshot, err
	}

	for rows.Next() {
		values := make([]any, len(snapshot.Columns))
		destinations := make([]any, len(values))
		for i := range values {
			destinations[i] = &values[i]
		}

		if err := rows.Scan(destinations...); err != nil {
			return snapshot, err
		}

		row := make([]string, len(values))
		for i, value := range values {
			switch value := value.(type) {
			case nil:
				row[i] = "NULL"
			case []byte:
				row[i] = string(value)
			default:
				row[i] = fmt.Sprint(value)
			}
		}
		snapshot.Rows = append(snapshot.Rows, row)
	}

	return snapshot, rows.Err()
}
