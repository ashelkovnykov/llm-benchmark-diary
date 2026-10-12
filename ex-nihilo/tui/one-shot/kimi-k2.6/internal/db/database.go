package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

type Manager struct {
	Dir string
}

func NewManager(dir string) *Manager {
	return &Manager{Dir: dir}
}

func (m *Manager) ListDBs() ([]string, error) {
	entries, err := os.ReadDir(m.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var dbs []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".db" {
			dbs = append(dbs, strings.TrimSuffix(e.Name(), ".db"))
		}
	}
	sort.Strings(dbs)
	return dbs, nil
}

func (m *Manager) DBPath(name string) string {
	return filepath.Join(m.Dir, name+".db")
}

func (m *Manager) Exists(name string) bool {
	_, err := os.Stat(m.DBPath(name))
	return err == nil
}

func (m *Manager) ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if m.Exists(name) {
		return fmt.Errorf("database '%s' already exists", name)
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, name)
	if !matched {
		return fmt.Errorf("name must contain only alphanumeric characters")
	}
	return nil
}

func (m *Manager) CreateDB(name string) (*sql.DB, error) {
	if err := os.MkdirAll(m.Dir, 0755); err != nil {
		return nil, err
	}
	if err := m.ValidateName(name); err != nil {
		return nil, err
	}

	dbPath := m.DBPath(name)
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	schemaBytes, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		database.Close()
		return nil, err
	}

	if _, err := database.Exec(string(schemaBytes)); err != nil {
		database.Close()
		return nil, err
	}

	if err := seedDummyData(database); err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}

func (m *Manager) OpenDB(name string) (*sql.DB, error) {
	dbPath := m.DBPath(name)
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, err
	}
	return database, nil
}

func seedDummyData(database *sql.DB) error {
	stmt, err := database.Prepare("INSERT INTO records (name, value) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= 15; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("Record %d", i), fmt.Sprintf("Value %d", i)); err != nil {
			return err
		}
	}
	return nil
}

type Record struct {
	RowID int64
	Name  string
	Value string
}

func GetStats(database *sql.DB) (int, error) {
	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM records").Scan(&count)
	return count, err
}

func GetRecords(database *sql.DB, limit int) ([]Record, error) {
	rows, err := database.Query("SELECT rowid, name, value FROM records LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Name, &r.Value); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
