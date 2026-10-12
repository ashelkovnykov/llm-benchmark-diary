package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type column struct {
	name string
	typ  string
}

// insertableColumns returns the columns of the records table that should be
// populated explicitly: everything except rowid and columns with defaults.
func insertableColumns(conn *sql.DB) ([]column, error) {
	rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", quote(RecordsTable)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []column
	for rows.Next() {
		var (
			cid, notNull, pk int
			name, typ        string
			dflt             sql.NullString
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			return nil, err
		}
		if strings.EqualFold(name, "rowid") || dflt.Valid {
			continue
		}
		cols = append(cols, column{name: name, typ: strings.ToUpper(typ)})
	}
	return cols, rows.Err()
}

// dummyValue produces a value matching SQLite's type affinity rules.
func dummyValue(c column, i int) any {
	switch {
	case strings.Contains(c.typ, "INT"):
		return i
	case strings.Contains(c.typ, "CHAR"),
		strings.Contains(c.typ, "CLOB"),
		strings.Contains(c.typ, "TEXT"):
		return fmt.Sprintf("%s_%d", c.name, i)
	case strings.Contains(c.typ, "BLOB"), c.typ == "":
		return []byte(fmt.Sprintf("%s_%d", c.name, i))
	case strings.Contains(c.typ, "REAL"),
		strings.Contains(c.typ, "FLOA"),
		strings.Contains(c.typ, "DOUB"):
		return float64(i) * 1.25
	default:
		return i
	}
}

func seed(conn *sql.DB, n int) error {
	cols, err := insertableColumns(conn)
	if err != nil {
		return err
	}

	var stmt string
	if len(cols) == 0 {
		stmt = fmt.Sprintf("INSERT INTO %s DEFAULT VALUES", quote(RecordsTable))
	} else {
		names := make([]string, len(cols))
		marks := make([]string, len(cols))
		for i, c := range cols {
			names[i] = quote(c.name)
			marks[i] = "?"
		}
		stmt = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			quote(RecordsTable), strings.Join(names, ", "), strings.Join(marks, ", "))
	}

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ps, err := tx.Prepare(stmt)
	if err != nil {
		return err
	}
	defer ps.Close()

	for i := 1; i <= n; i++ {
		args := make([]any, len(cols))
		for j, c := range cols {
			args[j] = dummyValue(c, i)
		}
		if _, err := ps.Exec(args...); err != nil {
			return err
		}
	}
	return tx.Commit()
}
