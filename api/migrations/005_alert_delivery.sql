CREATE TABLE alert_channels (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 type TEXT NOT NULL CHECK(type IN ('smtp','webhook')),
 name TEXT NOT NULL CHECK(length(name) BETWEEN 1 AND 200),
 enabled BOOLEAN NOT NULL DEFAULT FALSE,
 config_encrypted BYTEA NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 deleted_at TIMESTAMPTZ
);
CREATE TABLE alert_deliveries (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 alert_event_id UUID NOT NULL REFERENCES alert_events(id),
 channel_id UUID NOT NULL REFERENCES alert_channels(id),
 channel_type TEXT NOT NULL CHECK(channel_type IN ('smtp','webhook')),
 config_encrypted BYTEA NOT NULL,
 payload JSONB NOT NULL,
 status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','sent','failed','cancelled')),
 attempt_count INTEGER NOT NULL DEFAULT 0 CHECK(attempt_count BETWEEN 0 AND 5),
 last_error TEXT,
 next_attempt_at TIMESTAMPTZ DEFAULT clock_timestamp(),
 last_attempt_at TIMESTAMPTZ,
 accepted_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(alert_event_id,channel_id),
 CHECK((status='pending')=(next_attempt_at IS NOT NULL))
);
CREATE INDEX deliveries_due_idx ON alert_deliveries(next_attempt_at,id) WHERE status='pending';
CREATE INDEX deliveries_event_idx ON alert_deliveries(alert_event_id,created_at DESC,id DESC);
CREATE INDEX deliveries_channel_idx ON alert_deliveries(channel_id) WHERE status='pending';
