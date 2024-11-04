-- +goose Up
-- +goose StatementBegin
-- Up Migration
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(64) PRIMARY KEY,
    data BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    user_id INTEGER,
    last_accessed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    
    CONSTRAINT sessions_expires_check CHECK (expires_at > created_at),
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES controllers(id)
        ON DELETE CASCADE
);

-- Create regular index for expires_at (without predicate)
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Create regular combined index for user lookups
CREATE INDEX idx_sessions_user ON sessions(user_id, expires_at);

-- Create index for last accessed
CREATE INDEX idx_sessions_last_accessed ON sessions(last_accessed_at);

-- Optional: Create function to clean expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM sessions
    WHERE expires_at < CURRENT_TIMESTAMP
    RETURNING COUNT(*) INTO deleted_count;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS cleanup_expired_sessions();
DROP INDEX IF EXISTS idx_sessions_last_accessed;
DROP INDEX IF EXISTS idx_sessions_user;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
