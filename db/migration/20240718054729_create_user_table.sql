-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users 
(
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) UNIQUE NOT NULL,
    pass_hash TEXT UNIQUE NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);

CREATE TABLE IF NOT EXISTS apps
(
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    secret TEXT UNIQUE NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS apps;
-- +goose StatementEnd
