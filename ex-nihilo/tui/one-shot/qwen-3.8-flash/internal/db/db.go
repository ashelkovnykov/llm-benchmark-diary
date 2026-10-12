package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "embed"
	_ "modernc.org/sqlite"
)

var (
	alphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	ErrNameEmpty   = errors.New("database name cannot be empty")
	ErrNameInvalid = errors.New("database name may only contain letters and numbers")
	ErrNameExists  = errors.New("database name already exists")
)

//go:embed schema.sql
var schema string

type Info struct {
	Name string
	Path string
}

type Preview struct {
	Columns []string
	Rows    []string
}

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func List(dir string) ([]Info, error) {
	if err := EnsureDir(dir); err != nil {
		return nil, err
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.db"))
	if err != nil {
		return nil, err
	}

	infos := make([]Info, 0, len(matches))
	seen := make(map[string]bool, len(matches))

	for _, path := range matches {
		fi, err := os.Stat(path)
		if err != nil || fi.IsDir() {
			continue
		}

		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if name == "" || seen[name] {
			continue
		}

		seen[name] = true
		infos = append(infos, Info{Name: name, Path: path})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos, nil
}

func Exists(dir, name string) (bool, error) {
	_, err := os.Stat(filepath.Join(dir, name+".db"))
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func ValidateName(dir, name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameEmpty
	}

	if !alphaNumeric.MatchString(name) {
		return ErrNameInvalid
	}

	exists, err := Exists(dir, name)
	if err != nil {
		return err
	}

	if exists {
		return ErrNameExists
	}

	return nil
}

func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}

	conn.SetMaxOpenConns(1)

	return conn, nil
}

func Create(dir, name string) (*sql.DB, error) {
	if err := ValidateName(dir, name); err != nil {
		return nil, err
	}

	if err := EnsureDir(dir); err != nil {
		return nil, err
	}

	path := filepath.Join(dir, name+".db")

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}

	if err := PopulateDummyRecords(conn); err != nil {
		conn.Close()
		return nil, err
	}

	if err := conn.Close(); err != nil {
		return nil, err
	}

	return Open(path)
}

func PopulateDummyRecords(conn *sql.DB) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare("INSERT INTO records (rowid) VALUES (?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	for i := int64(1); i <= 15; i++ {
		if _, err := stmt.Exec(i); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func CountRecords(conn *sql.DB) (int64, error) {
	var count int64

	if err := conn.QueryRow("SELECT COUNT(*) FROM records").Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func QueryPreview(conn *sql.DB, limit int) (Preview, error) {
	rows, err := conn.Query("SELECT * FROM records ORDER BY rowid LIMIT ?", limit)
	if err != nil {
		return Preview{}, err
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return Preview{}, err
	}

	result := Preview{
		Columns: columns,
	}

	for rows.Next() {
		values := make([]any, len(columns))
		addresses := make([]any, len(columns))

		for i := range values {
			addresses[i] = &values[i]
		}

		if err := rows.Scan(addresses...); err != nil {
			return result, err
		}

		row := make([]string, len(columns))

		for i, value := range values {
			row[i] = valueToString(value)
		}

		result.Rows = append(result.Rows, strings.Join(row, " | "))
	}

	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

func valueToString(value any) string {
	switch v := value.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
