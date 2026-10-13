package storage

import (
	"context"
	"github.com/franklinlem/certguard/api/internal/alerts"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/testutil"
	"testing"
	"time"
)

func TestPostgresMetricsTransactionsAndRetention(t *testing.T) {
	s := deliveryStore(t)
	ctx := context.Background()
	activeChannel(t, s)
	e := newEndpoint(t, s, "example.com")
	first := checkAlert(t, s, e, "warning", 20, "a")
	m, err := s.Metrics(ctx)
	if err != nil || m.Checks != 1 || m.CheckFailures != 0 || m.Endpoints != 1 || m.Warning != 1 || m.PendingDeliveries != 1 || m.Buckets[7] != 1 {
		t.Fatal(m, err)
	}
	due(t, s, e.ID)
	if _, err = s.Check(ctx, e.ID, func(context.Context, Endpoint) certificate.Result {
		r := resultFor(e, "error", 0, "")
		r.Code = "dns_error"
		r.Certificate = nil
		r.DurationMS = 600
		return r
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.DeliverOne(ctx, func(context.Context, string, alerts.ChannelConfig, []byte, string) alerts.Outcome {
		return alerts.Outcome{Accepted: true}
	}); err != nil {
		t.Fatal(err)
	}
	m, err = s.Metrics(ctx)
	if err != nil || m.Checks != 2 || m.CheckFailures != 1 || m.DurationMS != 600 || m.Buckets[0] != 1 || m.Buckets[2] != 2 || m.Deliveries != 1 || m.DeliveryAccepted != 1 || m.PendingDeliveries != 1 {
		t.Fatal(m, err)
	}
	// Provoca rollback depois da inserção do resultado/evento/fila.
	if _, err = s.Pool.Exec(ctx, "ALTER TABLE operational_metrics ADD CONSTRAINT reject_test CHECK(checks_total<3)"); err != nil {
		t.Fatal(err)
	}
	due(t, s, e.ID)
	if _, err = s.Check(ctx, e.ID, func(context.Context, Endpoint) certificate.Result { return resultFor(e, "healthy", 90, "b") }); err == nil {
		t.Fatal("contador rejeitado não reverteu operação")
	}
	after, err := s.Metrics(ctx)
	if err != nil || after.Checks != 2 || after.PendingDeliveries != 1 || after.Errors != 1 {
		t.Fatal(after, err)
	}
	if _, err = s.Pool.Exec(ctx, "UPDATE certificate_checks SET checked_at=now()-interval '200 days' WHERE id=$1", first.ID); err != nil {
		t.Fatal(err)
	}
	if n, err := s.PruneHistory(ctx, 180, 100); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	retained, err := s.Metrics(ctx)
	if err != nil || retained.Checks != 2 || retained.Buckets[7] != 2 {
		t.Fatal("retenção reduziu contador", retained, err)
	}
}
func TestPostgresReadiness(t *testing.T) {
	s := &Store{Pool: testutil.Database(t)}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Ready(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Pool.Exec(ctx, "DELETE FROM schema_migrations WHERE version=1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Ready(ctx); err == nil {
		t.Fatal("schema incompleto pronto")
	}
	if _, err := s.Pool.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES(0)"); err != nil {
		t.Fatal(err)
	}
	if err := s.Ready(ctx); err == nil {
		t.Fatal("schema incompatível pronto")
	}
}
