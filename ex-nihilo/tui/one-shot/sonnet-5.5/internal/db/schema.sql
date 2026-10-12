-- Placeholder schema. Must contain a "records" table keyed by "rowid".
CREATE TABLE records (
    rowid      INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    category   TEXT NOT NULL,
    value      REAL NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
