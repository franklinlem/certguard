// Package migrations mantém migrações versionadas e transacionais.
package migrations

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed 001_initial.sql
var initial string

//go:embed 002_auth.sql
var auth string

//go:embed 003_scheduler.sql
var scheduler string

//go:embed 004_alert_events.sql
var events string

//go:embed 005_alert_delivery.sql
var delivery string

func Up(ctx context.Context, pool *pgxpool.Pool) error {
	tx, e := pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(724319001)"); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())"); e != nil {
		return e
	}
	for i, script := range []string{initial, auth, scheduler, events, delivery} {
		var exists bool
		if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", i+1).Scan(&exists); e != nil {
			return e
		}
		if !exists {
			if _, e = tx.Exec(ctx, script); e != nil {
				return fmt.Errorf("migração %03d: %w", i+1, e)
			}
			if _, e = tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES($1)", i+1); e != nil {
				return e
			}
		}
	}

	return tx.Commit(ctx)
}
