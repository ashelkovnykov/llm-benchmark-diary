-- ./internal/db/schema.sql
CREATE TABLE IF NOT EXISTS records (
    rowid INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NOT NULL,
    value TEXT NOT NULL,
    created_at TEXT NOT NULL
);
