-- Placeholder schema. Guaranteed by the project spec to be valid and to
-- contain a `records` table keyed on `rowid`.
CREATE TABLE records (
    rowid INTEGER PRIMARY KEY,
    name  TEXT NOT NULL,
    value TEXT NOT NULL
);
