CREATE TABLE records (
    rowid      INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL,
    value      REAL    NOT NULL,
    created_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);
