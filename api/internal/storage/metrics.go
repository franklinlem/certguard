package storage

import (
	"context"
	"errors"
	"github.com/franklinlem/certguard/api/internal/telemetry"
	"github.com/franklinlem/certguard/api/migrations"
	"github.com/jackc/pgx/v5"
)

func (s *Store) Ping(ctx context.Context) error { return s.Pool.Ping(ctx) }
func (s *Store) Ready(ctx context.Context) error {
	if err := s.Ping(ctx); err != nil {
		return err
	}
	var count, minVersion, maxVersion int
	if err := s.Pool.QueryRow(ctx, "SELECT count(*),coalesce(min(version),0),coalesce(max(version),0) FROM schema_migrations").Scan(&count, &minVersion, &maxVersion); err != nil {
		return err
	}
	if count != migrations.Version || minVersion != 1 || maxVersion != migrations.Version {
		return errors.New("schema incompatível")
	}
	return nil
}

func recordCheckMetrics(ctx context.Context, tx pgx.Tx, duration int64, failed bool) error {
	tag, err := tx.Exec(ctx, `UPDATE operational_metrics SET checks_total=checks_total+1,
 check_failures=check_failures+$2::boolean::int,duration_ms_sum=duration_ms_sum+$1,
 duration_buckets=ARRAY[duration_buckets[1]+($1<=100)::int,duration_buckets[2]+($1<=500)::int,
 duration_buckets[3]+($1<=1000)::int,duration_buckets[4]+($1<=2500)::int,
 duration_buckets[5]+($1<=5000)::int,duration_buckets[6]+($1<=10000)::int,
 duration_buckets[7]+($1<=30000)::int,duration_buckets[8]+1] WHERE singleton`, duration, failed)
	if err == nil && tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return err
}
func recordDeliveryMetrics(ctx context.Context, tx pgx.Tx, accepted bool) error {
	tag, err := tx.Exec(ctx, "UPDATE operational_metrics SET deliveries_total=deliveries_total+1,delivery_accepted=delivery_accepted+$1::boolean::int,delivery_failures=delivery_failures+(NOT $1)::int WHERE singleton", accepted)
	if err == nil && tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return err
}
func (s *Store) Metrics(ctx context.Context) (telemetry.Snapshot, error) {
	var m telemetry.Snapshot
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return m, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	err = tx.QueryRow(ctx, "SELECT checks_total,check_failures,duration_ms_sum,duration_buckets,deliveries_total,delivery_failures,delivery_accepted FROM operational_metrics WHERE singleton").Scan(&m.Checks, &m.CheckFailures, &m.DurationMS, &m.Buckets, &m.Deliveries, &m.DeliveryFailures, &m.DeliveryAccepted)
	if err != nil {
		return m, err
	}
	err = tx.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE e.enabled),count(*) FILTER(WHERE NOT e.enabled),
 count(*) FILTER(WHERE latest.status='healthy'),count(*) FILTER(WHERE latest.status='warning'),count(*) FILTER(WHERE latest.status='critical'),count(*) FILTER(WHERE latest.status='error'),count(*) FILTER(WHERE latest.status IS NULL),
 count(*) FILTER(WHERE (latest.metadata->'certificate'->>'not_after')::timestamptz>=now() AND (latest.metadata->'certificate'->>'not_after')::timestamptz<=now()+interval '30 days'),
 count(*) FILTER(WHERE e.enabled AND e.next_check_at<=now()) FROM monitored_endpoints e`+latestJoin+`WHERE e.deleted_at IS NULL`).Scan(&m.Endpoints, &m.Enabled, &m.Paused, &m.Healthy, &m.Warning, &m.Critical, &m.Errors, &m.Unchecked, &m.Expiring30, &m.Due)
	if err != nil {
		return m, err
	}
	err = tx.QueryRow(ctx, `SELECT count(*) FILTER(WHERE status='pending'),count(*) FILTER(WHERE status='failed'),coalesce(greatest(extract(epoch FROM now()-min(created_at) FILTER(WHERE status='pending')),0),0),coalesce(extract(epoch FROM max(last_attempt_at)),0) FROM alert_deliveries`).Scan(&m.PendingDeliveries, &m.FailedDeliveries, &m.OldestPendingSeconds, &m.LastDeliveryUnix)
	if err != nil {
		return m, err
	}
	if err = tx.QueryRow(ctx, "SELECT coalesce(extract(epoch FROM max(last_check_at)),0) FROM monitored_endpoints").Scan(&m.LastCheckUnix); err != nil {
		return m, err
	}
	return m, tx.Commit(ctx)
}
