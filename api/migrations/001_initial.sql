-- Migração transacional aplicada explicitamente por certguard migrate.
CREATE TABLE monitored_endpoints (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 200),
 hostname TEXT NOT NULL CHECK (hostname = lower(hostname) AND length(hostname) BETWEEN 3 AND 253),
 port INTEGER NOT NULL DEFAULT 443 CHECK (port = 443),
 enabled BOOLEAN NOT NULL DEFAULT TRUE,
 interval_minutes INTEGER NOT NULL DEFAULT 360 CHECK (interval_minutes >= 5),
 warning_days INTEGER NOT NULL DEFAULT 30,
 critical_days INTEGER NOT NULL DEFAULT 7,
 last_check_at TIMESTAMPTZ,
 next_check_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 CHECK (warning_days > critical_days AND critical_days >= 0),
 UNIQUE(hostname, port)
);
CREATE INDEX endpoints_due_idx ON monitored_endpoints(next_check_at) WHERE enabled;
CREATE TABLE certificate_checks (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 endpoint_id UUID NOT NULL REFERENCES monitored_endpoints(id),
 checked_at TIMESTAMPTZ NOT NULL,
 status TEXT NOT NULL CHECK (status IN ('healthy','warning','critical','error')),
 code TEXT NOT NULL,
 duration_ms BIGINT NOT NULL CHECK(duration_ms >= 0),
 resolved_ip INET,
 metadata JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX checks_endpoint_time_idx ON certificate_checks(endpoint_id, checked_at DESC);
CREATE INDEX checks_status_time_idx ON certificate_checks(status, checked_at DESC);
