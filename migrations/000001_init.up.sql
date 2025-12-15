-- Initial schema for Root Server

-- Services table
CREATE TABLE IF NOT EXISTS services (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    endpoints TEXT[] NOT NULL,
    capabilities TEXT[] NOT NULL,
    metadata JSONB,
    status VARCHAR(50) NOT NULL,
    registered_at TIMESTAMP NOT NULL,
    last_heartbeat TIMESTAMP NOT NULL,
    health_check_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_services_name ON services(name);
CREATE INDEX idx_services_status ON services(status);

-- Sessions table (if using PostgreSQL for sessions)
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    service_id VARCHAR(255) NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Token blacklist (for revoked tokens)
CREATE TABLE IF NOT EXISTS token_blacklist (
    token_id VARCHAR(255) PRIMARY KEY,
    revoked_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist(expires_at);

-- Audit log
CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    service_id VARCHAR(255),
    user_id VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    metadata JSONB,
    request_id VARCHAR(255)
);

CREATE INDEX idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_log_service_id ON audit_log(service_id);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
