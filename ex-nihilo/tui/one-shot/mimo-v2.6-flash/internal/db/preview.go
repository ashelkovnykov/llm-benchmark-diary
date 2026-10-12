package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// Preview is a lightweight snapshot of the records table.
type Preview struct {
	Total   int64
	Columns []string
	Rows    [][]string
}

// Preview returns the total record count plus the first limit rows.
func Preview(conn *sql.DB, limit int) (*Preview, error) {
	if limit <= 0 {
		limit = PreviewLimit
	}
	p := &Preview{}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM records`).Scan(&p.Total); err != nil {
		return nil, fmt.Errorf("count records: %w", err)
	}
	rows, err := conn.Query(`SELECT * FROM records ORDER BY rowid LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query records: %w", err)
	}
	defer rows.Close()

	if p.Columns, err = rows.Columns(); err != nil {
		return nil, err
	}
	values := make([]any, len(p.Columns))
	ptrs := make([]any, len(p.Columns))
	for i := range values {
		ptrs[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make([]string, len(p.Columns))
		for i, v := range values {
			row[i] = formatValue(v)
		}
		p.Rows = append(p.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return p, nil
}

func formatValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(t)
	case string:
		return t
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64)
	case time.Time:
		return t.Format(time.RFC3339)
	default:
		return fmt.Sprint(t)
	}
}
