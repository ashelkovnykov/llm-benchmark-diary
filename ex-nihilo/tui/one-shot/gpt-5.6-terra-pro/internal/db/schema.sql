-- Placeholder schema. Replaced later by the real one.
-- Guarantees: a table named "records" keyed by an explicit "rowid" column.

CREATE TABLE IF NOT EXISTS records (
    rowid      INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    value      INTEGER NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
