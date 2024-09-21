CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    password TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    provider VARCHAR(50) NOT NULL DEFAULT 'password',
    provider_id VARCHAR(255),
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (provider, provider_id),
    CONSTRAINT chk_authentication_method CHECK (
        (provider = 'password' AND password IS NOT NULL AND provider_id IS NULL) OR
        (provider != 'password' AND password IS NULL AND provider_id IS NOT NULL)
    )
);