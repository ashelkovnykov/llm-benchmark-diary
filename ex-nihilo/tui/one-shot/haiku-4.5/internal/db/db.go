package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

const dbDir = "./db"

func init() {
	os.MkdirAll(dbDir, 0755)
}

func ListDatabases() ([]string, error) {
	entries, err := os.ReadDir(dbDir)
	if err != nil {
		return nil, err
	}

	var dbs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			dbs = append(dbs, entry.Name()[:len(entry.Name())-3])
		}
	}
	return dbs, nil
}

func ValidateDatabaseName(name string) (bool, string) {
	if name == "" {
		return false, "Database name cannot be empty"
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(name) {
		return false, "Database name must contain only alphanumeric characters"
	}

	dbs, err := ListDatabases()
	if err == nil {
		for _, db := range dbs {
			if db == name {
				return false, "Database with this name already exists"
			}
		}
	}

	return true, ""
}

func CreateDatabase(name string) (*sql.DB, error) {
	dbPath := filepath.Join(dbDir, name+".db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	schema, err := os.ReadFile("./internal/db/schema.sql")
	if err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, err
	}

	if err := populateDummyData(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func OpenDatabase(name string) (*sql.DB, error) {
	dbPath := filepath.Join(dbDir, name+".db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func GetRecordCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM records").Scan(&count)
	return count, err
}

func GetFirstNRecords(db *sql.DB, n int) ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT * FROM records LIMIT ?", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		record := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				record[col] = string(b)
			} else {
				record[col] = val
			}
		}
		results = append(results, record)
	}

	return results, rows.Err()
}

func populateDummyData(db *sql.DB) error {
	stmt, err := db.Prepare("INSERT INTO records (name, value) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= 15; i++ {
		_, err := stmt.Exec(fmt.Sprintf("Record %d", i), fmt.Sprintf("Value %d", i))
		if err != nil {
			return err
		}
	}

	return nil
}
