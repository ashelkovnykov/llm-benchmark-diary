// ./internal/db/db.go
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

var alnumRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func EnsureDBDir() error {
	return os.MkdirAll("./db", 0755)
}

func ListDBs() ([]string, error) {
	if err := EnsureDBDir(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir("./db")
	if err != nil {
		return nil, err
	}
	var dbs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".db") || strings.HasSuffix(name, ".sqlite") || strings.HasSuffix(name, ".sqlite3") {
			dbs = append(dbs, name)
		}
	}
	return dbs, nil
}

func ValidateName(name string, existing []string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !alnumRe.MatchString(name) {
		return fmt.Errorf("name must be alphanumeric only")
	}
	for _, e := range existing {
		base := strings.TrimSuffix(e, filepath.Ext(e))
		if base == name {
			return fmt.Errorf("a DB named %q already exists", name)
		}
	}
	return nil
}

func CreateDB(name string) (string, error) {
	if err := EnsureDBDir(); err != nil {
		return "", err
	}
	path := filepath.Join("./db", name+".db")

	schemaBytes, err := os.ReadFile("./internal/db/schema.sql")
	if err != nil {
		return "", fmt.Errorf("read schema: %w", err)
	}

	// remove existing file if any (should not happen if validated)
	os.Remove(path)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if _, err := db.Exec(string(schemaBytes)); err != nil {
		return "", fmt.Errorf("exec schema: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	stmt, err := tx.Prepare(`INSERT INTO records (name, value, created_at) VALUES (?, ?, datetime('now'))`)
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	for i := 1; i <= 15; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("record-%02d", i), fmt.Sprintf("value-%02d", i)); err != nil {
			tx.Rollback()
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return path, nil
}

type Stats struct {
	Path  string
	Count int
	Cols  []string
	Rows  [][]string
}

func LoadStats(path string) (Stats, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return Stats{}, err
	}
	defer db.Close()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&count); err != nil {
		return Stats{}, fmt.Errorf("count: %w", err)
	}

	rows, err := db.Query(`SELECT rowid, name, value, created_at FROM records ORDER BY rowid LIMIT 10`)
	if err != nil {
		return Stats{}, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return Stats{}, err
	}

	var data [][]string
	for rows.Next() {
		// generic scan
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return Stats{}, err
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			if v == nil {
				row[i] = "NULL"
			} else {
				switch x := v.(type) {
				case string:
					row[i] = x
				case []byte:
					row[i] = string(x)
				case int64:
					row[i] = fmt.Sprintf("%d", x)
				default:
					row[i] = fmt.Sprintf("%v", x)
				}
			}
		}
		data = append(data, row)
	}
	return Stats{Path: path, Count: count, Cols: cols, Rows: data}, nil
}
