-- +goose Up
-- +goose StatementBegin
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'controller_role') THEN
        CREATE TYPE controller_role AS ENUM ('admin', 'user', 'super');
    END IF;
END $$;
CREATE TABLE IF NOT EXISTS controllers (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    first_name TEXT NOT NULL 
        CHECK (LENGTH(TRIM(first_name)) > 0)  -- Prevent empty strings
        CHECK (first_name ~ '^[A-Za-zÀ-ÖØ-öø-ÿ\s\-''\.]+$'), -- Allow letters, spaces, hyphens, apostrophes, periods
    last_name TEXT NOT NULL 
        CHECK (LENGTH(TRIM(first_name)) > 0)  -- Prevent empty strings
        CHECK (first_name ~ '^[A-Za-zÀ-ÖØ-öø-ÿ\s\-''\.]+$'), -- Allow letters, spaces, hyphens, apostrophes, periods
    initials VARCHAR(2) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    facility_id INTEGER NOT NULL,
    role controller_role NOT NULL,
    
    -- Add foreign key constraint for facility_id
    CONSTRAINT fk_facility
        FOREIGN KEY (facility_id)
        REFERENCES facilities(id)
        ON DELETE CASCADE
);
CREATE INDEX idx_controllers_email ON controllers(email);
CREATE INDEX idx_controllers_facility ON controllers(facility_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_controllers_email;
DROP INDEX IF EXISTS idx_controllers_facility;
DROP TABLE IF EXISTS controllers;
DROP TYPE IF  EXISTS controller_role;
-- +goose StatementEnd
