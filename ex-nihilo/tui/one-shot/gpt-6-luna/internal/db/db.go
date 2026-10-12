package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema embed.FS

const directory = "./db"

func List() ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".db" {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".db"))
	}
	return names, nil
}

func Create(name string) (string, error) {
	path := filepath.Join(directory, name+".db")
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("a database named %q already exists", name)
	} else if !os.IsNotExist(err) {
		return "", err
	}

	schemaSQL, err := schema.ReadFile("schema.sql")
	if err != nil {
		return "", err
	}

	database, err := sql.Open("sqlite", path)
	if err != nil {
		return "", err
	}
	defer database.Close()

	if _, err := database.Exec(string(schemaSQL)); err != nil {
		return "", err
	}

	for i := 1; i <= 15; i++ {
		if _, err := database.Exec(
			"INSERT INTO records (name, value) VALUES (?, ?)",
			fmt.Sprintf("Record %02d", i),
			fmt.Sprintf("Sample value %02d", i),
		); err != nil {
			return "", err
		}
	}

	return path, nil
}

func Load(path string) (int, []string, [][]string, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return 0, nil, nil, err
	}
	defer database.Close()

	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM records").Scan(&count); err != nil {
		return 0, nil, nil, err
	}

	rows, err := database.Query("SELECT * FROM records ORDER BY rowid LIMIT 10")
	if err != nil {
		return 0, nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return 0, nil, nil, err
	}

	var data [][]string
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return 0, nil, nil, err
		}

		record := make([]string, len(columns))
		for i, value := range values {
			switch v := value.(type) {
			case []byte:
				record[i] = string(v)
			default:
				record[i] = fmt.Sprint(v)
			}
		}
		data = append(data, record)
	}

	return count, columns, data, rows.Err()
}
