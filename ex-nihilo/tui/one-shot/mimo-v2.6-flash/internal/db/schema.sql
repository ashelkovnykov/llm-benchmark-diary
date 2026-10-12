-- Placeholder schema. Replace with the real schema when it lands.
-- Guaranteed invariants: table "records", primary key column "rowid".

CREATE TABLE IF NOT EXISTS records (
    rowid      INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL,
    category   TEXT    NOT NULL,
    value      INTEGER NOT NULL DEFAULT 0,
    notes      TEXT    NOT NULL DEFAULT '',
    created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_records_category ON records (category);
