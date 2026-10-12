package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

const Dir = "./db"

// List returns the names (without extension) of available DBs.
func List() ([]string, error) {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(Dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) == ".db" {
			names = append(names, name[:len(name)-len(".db")])
		}
	}
	return names, nil
}

func path(name string) string {
	return filepath.Join(Dir, name+".db")
}

// Exists reports whether a DB with the given name exists.
func Exists(name string) bool {
	_, err := os.Stat(path(name))
	return err == nil
}

// Open opens an existing DB.
func Open(name string) (*sql.DB, error) {
	return sql.Open("sqlite", path(name))
}

// Create builds a new DB from the schema and seeds dummy records.
func Create(name string) (*sql.DB, error) {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return nil, err
	}
	conn, err := sql.Open("sqlite", path(name))
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	if err := seed(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func seed(conn *sql.DB) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO records (name, value, created) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	for i := 1; i <= 15; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("record-%d", i), i*10, now); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// Row is a generic representation of a DB row.
type Row struct {
	Values []string
}

// Snapshot holds a count and the first rows for display.
type Snapshot struct {
	Columns []string
	Rows    []Row
	Count   int
}

// Load reads a count and up to limit rows from the records table.
func Load(conn *sql.DB, limit int) (Snapshot, error) {
	var snap Snapshot

	if err := conn.QueryRow("SELECT COUNT(*) FROM records").Scan(&snap.Count); err != nil {
		return snap, err
	}

	rows, err := conn.Query(fmt.Sprintf("SELECT * FROM records LIMIT %d", limit))
	if err != nil {
		return snap, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return snap, err
	}
	snap.Columns = cols

	for rows.Next() {
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return snap, err
		}
		vals := make([]string, len(cols))
		for i, v := range raw {
			vals[i] = fmt.Sprintf("%v", v)
		}
		snap.Rows = append(snap.Rows, Row{Values: vals})
	}
	return snap, rows.Err()
}
