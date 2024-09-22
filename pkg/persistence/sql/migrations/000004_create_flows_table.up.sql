CREATE TABLE IF NOT EXISTS flows (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    state VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    issued_at TIMESTAMP WITH TIME ZONE NOT NULL,
    request_url TEXT,
    errors JSONB
);

CREATE INDEX idx_flows_expires_at ON flows(expires_at);