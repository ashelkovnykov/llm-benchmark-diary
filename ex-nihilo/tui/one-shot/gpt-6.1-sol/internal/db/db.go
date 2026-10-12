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

const extension = ".sqlite"

type Database struct {
	SQL  *sql.DB
	Name string
}

type Preview struct {
	Count   int64
	Columns []string
	Rows    [][]string
}

func List(dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), extension) {
			names = append(names, strings.TrimSuffix(entry.Name(), extension))
		}
	}
	sort.Strings(names)
	return names, nil
}

func ValidateName(dir, name string) error {
	if name == "" {
		return errors.New("name must not be empty")
	}
	if !validName.MatchString(name) {
		return errors.New("name must contain only ASCII letters and numbers")
	}

	_, err := os.Lstat(filepath.Join(dir, name+extension))
	switch {
	case err == nil:
		return errors.New("a database with that name already exists")
	case !errors.Is(err, os.ErrNotExist):
		return err
	default:
		return nil
	}
}

func Open(dir, name string) (*Database, error) {
	// Names from List may contain punctuation, but must not escape dir.
	if name == "" || filepath.Base(name) != name {
		return nil, errors.New("invalid database filename")
	}

	path, err := filepath.Abs(filepath.Join(dir, name+extension))
	if err != nil {
		return nil, err
	}

	uri := url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(path),
		RawQuery: "mode=rw",
	}
	conn, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &Database{SQL: conn, Name: name}, nil
}

func Create(dir, name string) (_ *Database, err error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	if err := ValidateName(dir, name); err != nil {
		return nil, err
	}

	path := filepath.Join(dir, name+extension)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, errors.New("a database with that name already exists")
		}
		return nil, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return nil, err
	}

	complete := false
	defer func() {
		if !complete {
			_ = os.Remove(path)
		}
	}()

	database, err := Open(dir, name)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !complete {
			_ = database.Close()
		}
	}()

	tx, err := database.SQL.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(schema); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	// Dummy data for the placeholder schema.
	for i := 1; i <= 15; i++ {
		_, err := tx.Exec(
			`INSERT INTO records (name, description) VALUES (?, ?)`,
			fmt.Sprintf("Record %02d", i),
			fmt.Sprintf("Dummy description %02d", i),
		)
		if err != nil {
			return nil, fmt.Errorf("insert dummy record: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	complete = true
	return database, nil
}

func (d *Database) Preview() (Preview, error) {
	var preview Preview

	tx, err := d.SQL.Begin()
	if err != nil {
		return preview, err
	}
	defer tx.Rollback()

	if err := tx.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&preview.Count); err != nil {
		return preview, err
	}

	rows, err := tx.Query(`SELECT * FROM records ORDER BY rowid LIMIT 10`)
	if err != nil {
		return preview, err
	}
	defer rows.Close()

	preview.Columns, err = rows.Columns()
	if err != nil {
		return preview, err
	}

	for rows.Next() {
		values := make([]any, len(preview.Columns))
		destinations := make([]any, len(values))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := rows.Scan(destinations...); err != nil {
			return preview, err
		}

		row := make([]string, len(values))
		for i, value := range values {
			switch value := value.(type) {
			case nil:
				row[i] = "NULL"
			case []byte:
				row[i] = fmt.Sprintf("0x%x", value)
			default:
				row[i] = fmt.Sprint(value)
			}
		}
		preview.Rows = append(preview.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return preview, err
	}
	if err := rows.Close(); err != nil {
		return preview, err
	}
	if err := tx.Commit(); err != nil {
		return preview, err
	}

	return preview, nil
}

func (d *Database) Close() error {
	return d.SQL.Close()
}
