-- Create table
CREATE TABLE IF NOT EXISTS events (
    id           TEXT PRIMARY KEY,
    payload      BYTEA NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create index
CREATE INDEX IF NOT EXISTS idx_events_processed_at 
ON events (processed_at);
