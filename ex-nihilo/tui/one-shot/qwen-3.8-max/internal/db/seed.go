package db

import (
	"database/sql"
	"fmt"
)

// Seed inserts n dummy records into the records table.
func Seed(db *sql.DB, n int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO records (name, value) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= n; i++ {
		name := fmt.Sprintf("record_%03d", i)
		value := fmt.Sprintf("value_%d", i)
		if _, err := stmt.Exec(name, value); err != nil {
			return err
		}
	}

	return tx.Commit()
}
