CREATE TABLE IF NOT EXISTS documents (
    id           TEXT    PRIMARY KEY,   -- это и есть unique_id, генерируется приложением (UUID)
    user_id      INTEGER NOT NULL REFERENCES users(id),
    bucket_name  TEXT    NOT NULL,
    key          TEXT    NOT NULL,
    filename     TEXT    NOT NULL DEFAULT '',
    content_type TEXT    NOT NULL DEFAULT '',
    size_bytes   INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE (bucket_name, key)
);