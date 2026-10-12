package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// ListDBFiles returns the base names (without extension) of all .db files in dir.
func ListDBFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".db") {
			files = append(files, name[:len(name)-3])
		}
	}
	return files, nil
}

// CreateDB creates a new SQLite database at filePath, applies the embedded schema,
// and populates it with 15 dummy records.
func CreateDB(filePath, schemaSQL string) error {
	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	names := []string{
		"Alpha", "Bravo", "Charlie", "Delta", "Echo",
		"Foxtrot", "Golf", "Hotel", "India", "Juliet",
		"Kilo", "Lima", "Mike", "November", "Oscar",
	}
	for i, n := range names {
		amount := rand.Float64() * 1000
		if _, err := db.Exec("INSERT INTO records (name, amount) VALUES (?, ?)", n, amount); err != nil {
			return fmt.Errorf("insert record %d: %w", i, err)
		}
	}
	return nil
}

// LoadDB opens the database at filePath, returns the total number of records and
// the first 10 rows (each row as a slice of string values: rowid, name, amount).
func LoadDB(filePath string) (count int, rows [][]string, err error) {
	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		return 0, nil, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	err = db.QueryRow("SELECT COUNT(*) FROM records").Scan(&count)
	if err != nil {
		return 0, nil, fmt.Errorf("count records: %w", err)
	}

	qrows, err := db.Query("SELECT rowid, name, amount FROM records ORDER BY rowid LIMIT 10")
	if err != nil {
		return count, nil, fmt.Errorf("query records: %w", err)
	}
	defer qrows.Close()

	for qrows.Next() {
		var rowid int64
		var name string
		var amount float64
		if err := qrows.Scan(&rowid, &name, &amount); err != nil {
			return count, rows, fmt.Errorf("scan row: %w", err)
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", rowid),
			name,
			fmt.Sprintf("%.2f", amount),
		})
	}
	return count, rows, qrows.Err()
}
