-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
 (
    id UUID PRIMARY KEY,
    name VARCHAR(255)   NOT NULL,
    email VARCHAR(255)  NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);

-- CREATE INDEX idx_users_email ON users(email);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
