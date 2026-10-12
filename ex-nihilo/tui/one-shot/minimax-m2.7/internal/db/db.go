package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DBPath = "./db"

func ListDatabases() ([]string, error) {
	entries, err := os.ReadDir(DBPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var dbs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			dbs = append(dbs, entry.Name())
		}
	}
	return dbs, nil
}

func LoadDatabase(name string) (*sql.DB, error) {
	return sql.Open("sqlite3", filepath.Join(DBPath, name+".db"))
}

func CreateDatabase(name string, schema string) (*sql.DB, error) {
	if err := os.MkdirAll(DBPath, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(DBPath, name+".db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func ValidateName(name string, existingDBs []string) (bool, string) {
	if name == "" {
		return false, "name cannot be empty"
	}

	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false, "name must contain only alphanumeric characters"
		}
	}

	for _, db := range existingDBs {
		if db == name+".db" {
			return false, "a database with this name already exists"
		}
	}

	return true, ""
}
