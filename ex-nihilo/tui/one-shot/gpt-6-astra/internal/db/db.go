package db

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

var validName = regexp.MustCompile(`^[A-Za-z0-9]+$`)

type Snapshot struct {
	Name    string
	Count   int64
	Columns []string
	Rows    [][]string
}

func List(directory string) ([]string, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".db") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func Create(directory, name string) (Snapshot, error) {
	if name == "" {
		return Snapshot{}, errors.New("name cannot be empty")
	}
	if !validName.MatchString(name) {
		return Snapshot{}, errors.New("name must contain only letters A–Z, a–z, and digits 0–9")
	}

	filename := name + ".db"
	path := filepath.Join(directory, filename)

	// Reserve the filename atomically; never overwrite an existing database.
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if errors.Is(err, os.ErrExist) {
		return Snapshot{}, fmt.Errorf("a database named %q already exists", name)
	}
	if err != nil {
		return Snapshot{}, err
	}

	success := false
	defer func() {
		if !success {
			_ = os.Remove(path)
		}
	}()

	if err := file.Close(); err != nil {
		return Snapshot{}, err
	}

	conn, err := open(path, "rw")
	if err != nil {
		return Snapshot{}, err
	}
	defer conn.Close()

	if err := initialize(conn); err != nil {
		return Snapshot{}, err
	}

	snapshot, err := preview(conn, filename)
	if err != nil {
		return Snapshot{}, err
	}

	success = true
	return snapshot, nil
}

func Load(directory, filename string) (Snapshot, error) {
	conn, err := open(filepath.Join(directory, filename), "ro")
	if err != nil {
		return Snapshot{}, err
	}
	defer conn.Close()

	return preview(conn, filename)
}

func open(path, mode string) (*sql.DB, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	uri := url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(absolute),
		RawQuery: "mode=" + mode,
	}

	conn, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		conn.Close()
		return nil, err
	}
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func initialize(conn *sql.DB) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	// Placeholder seed data, matching the placeholder schema.
	for i := 1; i <= 15; i++ {
		if _, err := tx.Exec(
			`INSERT INTO records (label) VALUES (?)`,
			fmt.Sprintf("Dummy record %02d", i),
		); err != nil {
			return fmt.Errorf("seed records: %w", err)
		}
	}

	return tx.Commit()
}

func preview(conn *sql.DB, filename string) (Snapshot, error) {
	result := Snapshot{Name: filename}

	tx, err := conn.Begin()
	if err != nil {
		return Snapshot{}, err
	}
	defer tx.Rollback()

	if err := tx.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&result.Count); err != nil {
		return Snapshot{}, fmt.Errorf("count records: %w", err)
	}

	rows, err := tx.Query(`SELECT * FROM records ORDER BY rowid LIMIT 10`)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read records: %w", err)
	}
	defer rows.Close()

	result.Columns, err = rows.Columns()
	if err != nil {
		return Snapshot{}, err
	}

	for rows.Next() {
		values := make([]any, len(result.Columns))
		targets := make([]any, len(values))
		for i := range values {
			targets[i] = &values[i]
		}

		if err := rows.Scan(targets...); err != nil {
			return Snapshot{}, err
		}

		row := make([]string, len(values))
		for i, value := range values {
			switch v := value.(type) {
			case nil:
				row[i] = "NULL"
			case []byte:
				row[i] = fmt.Sprintf("0x%x", v)
			default:
				row[i] = fmt.Sprint(v)
			}
		}
		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return Snapshot{}, err
	}
	if err := rows.Close(); err != nil {
		return Snapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return Snapshot{}, err
	}

	return result, nil
}
