package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

var alphanumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func DBDir() string {
	return "db"
}

func ListDBs() ([]string, error) {
	dir := DBDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var dbs []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".db") {
			name := strings.TrimSuffix(e.Name(), ".db")
			dbs = append(dbs, name)
		}
	}
	return dbs, nil
}

func ValidateName(name string, existing []string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Name cannot be empty", false
	}
	if !alphanumeric.MatchString(name) {
		return "Name must contain only alphanumeric characters", false
	}
	for _, e := range existing {
		if e == name {
			return "A database with that name already exists", false
		}
	}
	return "", true
}

func CreateDB(name string) (string, error) {
	dir := DBDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name+".db")
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return "", fmt.Errorf("read schema: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return "", err
	}
	defer db.Close()
	if _, err := db.Exec(string(schema)); err != nil {
		os.Remove(path)
		return "", fmt.Errorf("apply schema: %w", err)
	}
	// Insert 15 dummy records
	for i := 1; i <= 15; i++ {
		_, err := db.Exec("INSERT INTO records (description, quantity) VALUES (?, ?)",
			fmt.Sprintf("Dummy record %d", i), i*10)
		if err != nil {
			os.Remove(path)
			return "", fmt.Errorf("insert dummy: %w", err)
		}
	}
	return path, nil
}

func OpenDB(name string) (*sql.DB, string, error) {
	path := filepath.Join(DBDir(), name+".db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, "", err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, "", err
	}
	return db, path, nil
}

type Record struct {
	RowID       int64
	Description string
	Quantity    int
}

func GetStats(db *sql.DB) (int, []Record, error) {
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM records").Scan(&total)
	if err != nil {
		return 0, nil, err
	}
	rows, err := db.Query("SELECT rowid, description, quantity FROM records ORDER BY rowid LIMIT 10")
	if err != nil {
		return total, nil, err
	}
	defer rows.Close()
	var recs []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Description, &r.Quantity); err != nil {
			return total, recs, err
		}
		recs = append(recs, r)
	}
	return total, recs, rows.Err()
}

func FormatRecords(recs []Record) string {
	if len(recs) == 0 {
		return "(no rows)"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-6s %-20s %s\n", "rowid", "description", "quantity"))
	b.WriteString(strings.Repeat("-", 40) + "\n")
	for _, r := range recs {
		b.WriteString(fmt.Sprintf("%-6d %-20s %d\n", r.RowID, r.Description, r.Quantity))
	}
	return b.String()
}
