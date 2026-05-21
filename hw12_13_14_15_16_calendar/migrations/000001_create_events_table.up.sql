-- Create events table
CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    description TEXT,
    user_id VARCHAR(255) NOT NULL,
    notify_before BIGINT NOT NULL DEFAULT 0
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time);
CREATE INDEX IF NOT EXISTS idx_events_end_time ON events(end_time);