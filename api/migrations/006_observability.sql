CREATE TABLE operational_metrics (
 singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK(singleton),
 checks_total BIGINT NOT NULL DEFAULT 0,
 check_failures BIGINT NOT NULL DEFAULT 0,
 duration_ms_sum BIGINT NOT NULL DEFAULT 0,
 duration_buckets BIGINT[] NOT NULL CHECK(cardinality(duration_buckets)=8),
 deliveries_total BIGINT NOT NULL DEFAULT 0,
 delivery_failures BIGINT NOT NULL DEFAULT 0,
 delivery_accepted BIGINT NOT NULL DEFAULT 0
);
-- Inicializa apenas o histórico ainda disponível; não reconstrói registros removidos.
INSERT INTO operational_metrics
SELECT true,count(*),count(*) FILTER(WHERE code<>'success'),coalesce(sum(duration_ms),0),
 ARRAY[count(*) FILTER(WHERE duration_ms<=100),count(*) FILTER(WHERE duration_ms<=500),
 count(*) FILTER(WHERE duration_ms<=1000),count(*) FILTER(WHERE duration_ms<=2500),
 count(*) FILTER(WHERE duration_ms<=5000),count(*) FILTER(WHERE duration_ms<=10000),
 count(*) FILTER(WHERE duration_ms<=30000),count(*)],
 (SELECT coalesce(sum(attempt_count),0) FROM alert_deliveries),
 (SELECT coalesce(sum(attempt_count),0)-count(*) FILTER(WHERE status='sent') FROM alert_deliveries),
 (SELECT count(*) FROM alert_deliveries WHERE status='sent')
FROM certificate_checks;
CREATE INDEX checks_current_target_idx ON certificate_checks
 (endpoint_id,(metadata->>'hostname'),(metadata->>'port'),checked_at DESC,id DESC);
CREATE INDEX deliveries_created_idx ON alert_deliveries(created_at DESC,id DESC);
