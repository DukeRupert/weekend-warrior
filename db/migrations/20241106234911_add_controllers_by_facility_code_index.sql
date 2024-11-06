-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_facilities_code_id ON facilities(code, id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_facilities_code_id;
-- +goose StatementEnd
