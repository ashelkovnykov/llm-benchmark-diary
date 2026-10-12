package db

import (
	"database/sql"
	"fmt"
	"strings"
)

const dummyCount = 15

type column struct {
	name string
	typ  string
	pk   bool
}

// isAutoRowID reports whether the column is an auto-assigned integer key and
// should therefore be skipped when seeding.
func (c column) isAutoRowID() bool {
	return c.pk && strings.Contains(strings.ToUpper(c.typ), "INT")
}

// seedRecords inserts count dummy rows into the records table. Values are
// generated from each column's declared type, so the seeding logic keeps
// working when the schema changes.
func seedRecords(conn *sql.DB, count int) error {
	cols, err := tableColumns(conn)
	if err != nil {
		return err
	}

	var names, placeholders []string
	var seeded []column
	for _, c := range cols {
		if c.isAutoRowID() {
			continue
		}
		seeded = append(seeded, c)
		names = append(names, quoteIdent(c.name))
		placeholders = append(placeholders, "?")
	}
	if len(seeded) == 0 {
		return nil
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quoteIdent(TableName), strings.Join(names, ", "), strings.Join(placeholders, ", "))

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	for i := 1; i <= count; i++ {
		args := make([]any, len(seeded))
		for j, c := range seeded {
			args[j] = dummyValue(c, i)
		}
		if _, err := tx.Exec(query, args...); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func tableColumns(conn *sql.DB) ([]column, error) {
	rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", quoteIdent(TableName)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []column
	for rows.Next() {
		var (
			cid, notNull, pk int
			name, typ        string
			def              sql.NullString
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &def, &pk); err != nil {
			return nil, err
		}
		cols = append(cols, column{name: name, typ: typ, pk: pk > 0})
	}
	return cols, rows.Err()
}

func dummyValue(c column, i int) any {
	switch t := strings.ToUpper(c.typ); {
	case strings.Contains(t, "INT"):
		return i
	case strings.Contains(t, "REAL"), strings.Contains(t, "FLOA"), strings.Contains(t, "DOUB"):
		return float64(i) * 1.5
	case strings.Contains(t, "BLOB"):
		return []byte(fmt.Sprintf("%s_%d", c.name, i))
	default:
		return fmt.Sprintf("%s_%d", c.name, i)
	}
}
