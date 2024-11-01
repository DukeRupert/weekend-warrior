-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS controllers(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    initials CHAR(2) NOT NULL,
    email TEXT UNIQUE NOT NULL,
    facility_id INTEGER NOT NULL REFERENCES facilities(id),
    role_id INTEGER NOT NULL REFERENCES roles(id),
    UNIQUE(facility_id, initials)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS controllers;
-- +goose StatementEnd
