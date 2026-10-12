CREATE TABLE IF NOT EXISTS records (
    rowid      INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL,
    category   TEXT,
    quantity   INTEGER NOT NULL,
    price      REAL,
    created_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);
