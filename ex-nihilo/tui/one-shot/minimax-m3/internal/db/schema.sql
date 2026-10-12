CREATE TABLE IF NOT EXISTS records (
    rowid       INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT NOT NULL,
    description TEXT,
    amount      REAL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
