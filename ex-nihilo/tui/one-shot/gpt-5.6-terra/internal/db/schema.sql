-- Placeholder schema. Replace with the real one when available.
-- Requirements: valid SQLite, contains a table named "records",
-- and uses a column named "rowid" as the primary key.

CREATE TABLE IF NOT EXISTS records (
    rowid      INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL,
    value      INTEGER NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
