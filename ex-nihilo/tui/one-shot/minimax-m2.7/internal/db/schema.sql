CREATE TABLE records (
    rowid INTEGER PRIMARY KEY,
    name TEXT,
    description TEXT,
    value INTEGER
);

CREATE INDEX idx_records_rowid ON records(rowid);
