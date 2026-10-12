-- Placeholder schema. Replace with the real schema when available.
-- Requirements: a table named "records" whose primary key column is "rowid".

CREATE TABLE IF NOT EXISTS records (
    rowid      INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    category   TEXT    NOT NULL,
    value      INTEGER NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
