-- Placeholder schema. The real schema is guaranteed to define a table named
-- "records" with a primary key column named "rowid".
CREATE TABLE records (
    rowid  INTEGER PRIMARY KEY,
    name   TEXT    NOT NULL,
    email  TEXT    NOT NULL,
    score  REAL    NOT NULL,
    active INTEGER NOT NULL DEFAULT 0
);
