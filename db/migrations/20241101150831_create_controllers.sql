-- +goose Up
-- +goose StatementBegin
CREATE TYPE controller_role AS ENUM ('admin', 'user', 'super');
CREATE TABLE IF NOT EXISTS controllers (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    initials VARCHAR(10) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    facility_id INTEGER NOT NULL,
    role controller_role NOT NULL,
    
    -- Add foreign key constraint for facility_id
    CONSTRAINT fk_facility
        FOREIGN KEY (facility_id)
        REFERENCES facilities(id)
        ON DELETE RESTRICT
);
CREATE INDEX idx_controllers_email ON controllers(email);
CREATE INDEX idx_controllers_facility ON controllers(facility_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_controllers_email;
DROP INDEX IF EXISTS idx_controllers_facility;
DROP TABLE IF EXISTS controllers;
DROP TYPE IF EXISTS controller_role;
-- +goose StatementEnd
