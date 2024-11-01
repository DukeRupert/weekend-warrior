-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS controller_facility_roles(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    controller_id INTEGER NOT NULL REFERENCES controllers(id),
    facility_id INTEGER NOT NULL REFERENCES facilities(id),
    role_id INTEGER NOT NULL REFERENCES roles(id),
    -- Ensure a controller can't have multiple roles at the same facility
    UNIQUE(controller_id, facility_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS controller_facility_roles;
-- +goose StatementEnd
