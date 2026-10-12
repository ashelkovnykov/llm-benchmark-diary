package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema []byte

type Record struct {
	ID    int
	Count int
}

func Create(path string) (conn *sql.DB, err error) {
	conn, err = sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if _, err = conn.Exec(string(schema)); err != nil {
		_ = conn.Close()
		_ = os.Remove(path)
		return nil, fmt.Errorf("error applying schema: %w", err)
	}

	return
}

func Open(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", path)
}

func CountRecords(conn *sql.DB) (n int, err error) {
	err = conn.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&n)
	return
}

func FetchRecords(conn *sql.DB, limit int, offset int) ([]Record, error) {
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}

	const q = `
	    SELECT rowid, count
	      FROM records
	  ORDER BY rowid
	     LIMIT ?
	    OFFSET ?`

	rows, err := conn.Query(q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select records: %w", err)
	}
	defer rows.Close()

	records := make([]Record, 0, limit)
	for rows.Next() {
		var r Record
		if err = rows.Scan(&r.ID, &r.Count); err != nil {
			return nil, fmt.Errorf("scan record: %w", err)
		}
		records = append(records, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("select records: %w", err)
	}

	return records, nil
}
