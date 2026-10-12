CREATE TABLE records (
    rowid INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
