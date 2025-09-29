-- Drop the table if it exists to ensure a clean slate and prevent errors on re-running.
DROP TABLE IF EXISTS notification_logs;

-- Create the table for logging all notification attempts using PostgreSQL syntax.
CREATE TABLE notification_logs (
    id SERIAL PRIMARY KEY,
    recipient TEXT NOT NULL,
    template_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL, -- e.g., 'sent', 'failed'
    details TEXT,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
