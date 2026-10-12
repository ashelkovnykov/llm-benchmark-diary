-- Placeholder schema for dbnav.
--
-- The real schema is not yet known. This placeholder satisfies the two
-- guarantees the application relies on:
--   * it contains a table named "records"
--   * "records" uses a column named "rowid" as its primary key
--
-- Replace this file with the real schema when it is available; no Go
-- changes should be required as long as those two guarantees hold.

CREATE TABLE IF NOT EXISTS records (
    rowid      INTEGER PRIMARY KEY,
    title      TEXT NOT NULL,
    body       TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
