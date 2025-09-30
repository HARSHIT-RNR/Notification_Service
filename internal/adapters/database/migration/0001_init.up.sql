-- Create the table for logging all notification attempts using PostgreSQL syntax.
CREATE TABLE notification_mails (
    id SERIAL PRIMARY KEY,
    recipient TEXT NOT NULL,
    template_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL, -- e.g., 'sent', 'failed'
    details TEXT,
    data JSONB NOT NULL DEFAULT '{}',
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

