CREATE TABLE IF NOT EXISTS object (
    id INTEGER NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_last_seen ON object (last_seen);