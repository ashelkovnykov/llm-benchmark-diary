package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	dir    string
	schema string
	conn   *sql.DB
}

type Record struct {
	RowID     int64
	Name      string
	Value     string
	CreatedAt string
}

func NewStore(dir, schemaPath string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	return &Store{dir: dir, schema: string(schema)}, nil
}

func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var dbs []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".db") {
			dbs = append(dbs, strings.TrimSuffix(e.Name(), ".db"))
		}
	}
	return dbs, nil
}

var nameRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// ValidateName returns a non-empty reason string if the name is invalid.
func (s *Store) ValidateName(name string) string {
	if name == "" {
		return "name cannot be empty"
	}
	if !nameRe.MatchString(name) {
		return "name may only contain alphanumeric characters"
	}
	exists, err := s.Exists(name)
	if err != nil {
		return fmt.Sprintf("could not check existing databases: %v", err)
	}
	if exists {
		return "a database with that name already exists"
	}
	return ""
}

func (s *Store) Exists(name string) (bool, error) {
	_, err := os.Stat(s.path(name))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (s *Store) path(name string) string {
	return filepath.Join(s.dir, name+".db")
}

func (s *Store) Create(name string) error {
	path := s.path(name)
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.Exec(s.schema); err != nil {
		os.Remove(path)
		return err
	}
	if err := seed(conn); err != nil {
		return err
	}
	return conn.Close()
}

func (s *Store) Load(name string) error {
	conn, err := sql.Open("sqlite3", s.path(name))
	if err != nil {
		return err
	}
	if s.conn != nil {
		s.conn.Close()
	}
	s.conn = conn
	return nil
}

func (s *Store) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

// Count returns the number of rows in the records table.
func (s *Store) Count() (int, error) {
	var n int
	err := s.conn.QueryRow("SELECT COUNT(*) FROM records").Scan(&n)
	return n, err
}

// First returns up to limit rows ordered by rowid.
func (s *Store) First(limit int) ([]Record, error) {
	rows, err := s.conn.Query("SELECT rowid, name, value, created_at FROM records ORDER BY rowid LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recs []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Name, &r.Value, &r.CreatedAt); err != nil {
			return nil, err
		}
		recs = append(recs, r)
	}
	return recs, rows.Err()
}

func seed(conn *sql.DB) error {
	for i := 1; i <= 15; i++ {
		if _, err := conn.Exec(
			"INSERT INTO records (name, value) VALUES (?, ?)",
			fmt.Sprintf("record_%02d", i),
			fmt.Sprintf("dummy value %d", i),
		); err != nil {
			return err
		}
	}
	return nil
}
