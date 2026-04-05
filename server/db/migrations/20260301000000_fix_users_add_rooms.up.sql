-- Fix users table: unique username + timestamp
ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);
ALTER TABLE users ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Rooms table (metadata only — messages are NOT persisted by design)
CREATE TABLE rooms (
    id VARCHAR(8) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    creator_id BIGINT REFERENCES users(id),
    pin_hash VARCHAR(255),
    ttl_minutes INT NOT NULL DEFAULT 60,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_rooms_active_expires ON rooms (is_active, expires_at);
