CREATE TABLE alert_events (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 endpoint_id UUID NOT NULL REFERENCES monitored_endpoints(id),
 check_id UUID REFERENCES certificate_checks(id) ON DELETE SET NULL,
 endpoint_name TEXT NOT NULL,
 hostname TEXT NOT NULL,
 port INTEGER NOT NULL CHECK (port=443),
 event_type TEXT NOT NULL,
 severity TEXT NOT NULL CHECK (severity IN ('info','warning','critical')),
 threshold_days INTEGER NOT NULL DEFAULT 0,
 dedupe_key TEXT NOT NULL,
 snapshot JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 resolved_at TIMESTAMPTZ,
 resolution TEXT CHECK (resolution IN ('recovered','changed','target_changed','archived')),
 CHECK ((resolved_at IS NULL) = (resolution IS NULL))
);
CREATE UNIQUE INDEX alerts_one_active_idx ON alert_events(endpoint_id) WHERE resolved_at IS NULL;
CREATE INDEX alerts_created_idx ON alert_events(created_at DESC,id DESC);
CREATE INDEX alerts_endpoint_idx ON alert_events(endpoint_id,created_at DESC,id DESC);
CREATE INDEX alerts_check_idx ON alert_events(check_id) WHERE check_id IS NOT NULL;
