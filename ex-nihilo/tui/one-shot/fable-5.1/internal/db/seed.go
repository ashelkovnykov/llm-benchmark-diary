package db

import (
	"database/sql"
	"fmt"
)

// seed populates the placeholder schema with n dummy records.
// Replace when the real schema / import flow lands.
func seed(conn *sql.DB, n int) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO records (name, value) VALUES (?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i := 1; i <= n; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("record-%02d", i), float64(i)*1.5); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
