package migrations

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

// Executar somente em banco descartável: o teste usa schema isolado.
func TestPostgres(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL não definido")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, e := pgxpool.New(ctx, url)
	if e != nil {
		t.Fatal(e)
	}
	defer admin.Close()
	schema := "test_certguard_" + time.Now().Format("150405000000000")
	if _, e = admin.Exec(ctx, "CREATE SCHEMA "+schema); e != nil {
		t.Fatal(e)
	}
	defer func() { _, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()
	cfg, e := pgxpool.ParseConfig(url)
	if e != nil {
		t.Fatal(e)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	// Simula atualização de uma instalação da etapa 02 com registros existentes.
	if _, e = pool.Exec(ctx, initial); e != nil {
		t.Fatal(e)
	}
	if _, e = pool.Exec(ctx, auth); e != nil {
		t.Fatal(e)
	}
	if _, e = pool.Exec(ctx, `CREATE TABLE schema_migrations(version INTEGER PRIMARY KEY,applied_at TIMESTAMPTZ NOT NULL DEFAULT now()); INSERT INTO schema_migrations(version) VALUES(1),(2); INSERT INTO monitored_endpoints(name,hostname,enabled) VALUES('Legado','legacy.example.com',true),('Pausado','paused.example.com',false)`); e != nil {
		t.Fatal(e)
	}
	if _, e = pool.Exec(ctx, `INSERT INTO certificate_checks(endpoint_id,checked_at,status,code,duration_ms,metadata) SELECT id,now(),'healthy','success',250,'{}' FROM monitored_endpoints WHERE hostname='legacy.example.com'`); e != nil {
		t.Fatal(e)
	}
	for range 2 {
		if e = Up(ctx, pool); e != nil {
			t.Fatal(e)
		}
	}
	var metricChecks, sumMS int64
	var buckets []int64
	if e = pool.QueryRow(ctx, "SELECT checks_total,duration_ms_sum,duration_buckets FROM operational_metrics").Scan(&metricChecks, &sumMS, &buckets); e != nil || metricChecks != 1 || sumMS != 250 || buckets[0] != 0 || buckets[1] != 1 || buckets[7] != 1 {
		t.Fatal("seed de métricas", metricChecks, sumMS, buckets, e)
	}
	var scheduled, paused int
	if e = pool.QueryRow(ctx, "SELECT count(*) FILTER (WHERE enabled AND next_check_at IS NOT NULL),count(*) FILTER (WHERE NOT enabled AND next_check_at IS NULL) FROM monitored_endpoints").Scan(&scheduled, &paused); e != nil || scheduled != 1 || paused != 1 {
		t.Fatal("migração de agendas", scheduled, paused, e)
	}
	var id string
	if e = pool.QueryRow(ctx, "INSERT INTO monitored_endpoints(name,hostname) VALUES('Exemplo','example.com') RETURNING id").Scan(&id); e != nil {
		t.Fatal(e)
	}
	for _, q := range []string{"INSERT INTO monitored_endpoints(name,hostname,port) VALUES('a','other.com',80)", "INSERT INTO monitored_endpoints(name,hostname,interval_minutes) VALUES('a','other.com',1)", "INSERT INTO monitored_endpoints(name,hostname,warning_days,critical_days) VALUES('a','other.com',7,30)", "INSERT INTO monitored_endpoints(name,hostname) VALUES('a','example.com')"} {
		if _, e = pool.Exec(ctx, q); e == nil {
			t.Fatal("constraint ausente")
		}
	}
	if _, e = pool.Exec(ctx, "INSERT INTO certificate_checks(endpoint_id,checked_at,status,code,duration_ms,metadata) VALUES($1,now(),'healthy','success',1,'{}')", id); e != nil {
		t.Fatal(e)
	}
	var n int
	if e = pool.QueryRow(ctx, "SELECT count(*) FROM certificate_checks WHERE endpoint_id=$1", id).Scan(&n); e != nil || n != 1 {
		t.Fatal(e, n)
	}
}
