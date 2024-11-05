-- +goose Up
-- +goose StatementBegin
-- Up Migration
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,           -- unique session identifier
    user_id INTEGER NOT NULL,      -- reference to the authenticated user
    created_at TIMESTAMP NOT NULL, -- when session was created
    expires_at TIMESTAMP NOT NULL, -- when session should expire
    ip_address TEXT,              -- optional: for security tracking
    user_agent TEXT,              -- optional: for security tracking
    is_active BOOLEAN DEFAULT true -- optional: for manual invalidation
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
