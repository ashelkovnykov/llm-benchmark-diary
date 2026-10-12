package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Record struct {
	RowID int
	Name  string
	Value int
}

func GetDatabases() ([]string, error) {
	files, err := os.ReadDir("./db")
	if err != nil {
		return nil, err
	}
	var dbs []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".db") {
			dbs = append(dbs, f.Name())
		}
	}
	return dbs, nil
}

func CreateDatabase(name string) error {
	dbPath := filepath.Join(".", "db", name+".db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	schema, err := os.ReadFile("./internal/db/schema.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return err
	}

	// Insert 15 dummy records
	for i := 1; i <= 15; i++ {
		_, err = db.Exec("INSERT INTO records (name, value) VALUES (?, ?)", fmt.Sprintf("Dummy Record %d", i), i*100)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDatabaseStats(name string) (int, []Record, error) {
	dbPath := filepath.Join(".", "db", name)
	if !strings.HasSuffix(name, ".db") {
		dbPath += ".db"
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return 0, nil, err
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM records").Scan(&count)
	if err != nil {
		return 0, nil, err
	}

	rows, err := db.Query("SELECT rowid, name, value FROM records LIMIT 10")
	if err != nil {
		return count, nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Name, &r.Value); err != nil {
			return count, nil, err
		}
		records = append(records, r)
	}
	return count, records, nil
}
