package db

import (
	"database/sql"
	"fmt"
)

type Record struct {
	RowID       int64
	Name        string
	Description string
	Value       int64
}

func GetRecordCount(db *sql.DB) (int64, error) {
	var count int64
	err := db.QueryRow("SELECT COUNT(*) FROM records").Scan(&count)
	return count, err
}

func GetRecords(db *sql.DB, limit int) ([]Record, error) {
	rows, err := db.Query("SELECT rowid, name, description, value FROM records LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.RowID, &r.Name, &r.Description, &r.Value); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func InsertDummyRecords(db *sql.DB, count int) error {
	stmt, err := db.Prepare("INSERT INTO records (name, description, value) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= count; i++ {
		_, err := stmt.Exec(
			fmt.Sprintf("Record %d", i),
			fmt.Sprintf("This is dummy record number %d", i),
			i*10,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
