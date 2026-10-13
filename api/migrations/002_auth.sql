CREATE TABLE users (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 email TEXT NOT NULL UNIQUE CHECK(email=lower(email)),
 password_hash TEXT NOT NULL,
 role TEXT NOT NULL DEFAULT 'admin' CHECK(role='admin'),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE sessions (
 token_hash TEXT PRIMARY KEY,
 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 expires_at TIMESTAMPTZ NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX sessions_expiry_idx ON sessions(expires_at);
ALTER TABLE monitored_endpoints ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE monitored_endpoints DROP CONSTRAINT monitored_endpoints_hostname_port_key;
CREATE UNIQUE INDEX endpoints_active_hostname_idx ON monitored_endpoints(hostname,port) WHERE deleted_at IS NULL;
ALTER TABLE monitored_endpoints ADD CONSTRAINT endpoints_max_warning CHECK(warning_days <= 36500);
ALTER TABLE monitored_endpoints ADD CONSTRAINT endpoints_max_interval CHECK(interval_minutes <= 10080);
