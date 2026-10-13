UPDATE monitored_endpoints
SET next_check_at = CASE WHEN enabled AND deleted_at IS NULL
    THEN clock_timestamp() + random() * interval '30 seconds' ELSE NULL END;
DROP INDEX endpoints_due_idx;
CREATE INDEX endpoints_due_idx ON monitored_endpoints(next_check_at, id)
    WHERE enabled AND deleted_at IS NULL;
CREATE INDEX checks_retention_idx ON certificate_checks(checked_at, id);
