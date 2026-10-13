package storage

import (
	"context"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/testutil"
	"testing"
	"time"
)

func resultFor(e Endpoint, status string, days int, fingerprint string) certificate.Result {
	yes := true
	return certificate.Result{Hostname: e.Hostname, Port: e.Port, Code: "success", Status: status, CheckedAt: time.Now().UTC(), DaysRemaining: &days, HostnameValid: &yes, ChainValid: &yes, Certificate: &certificate.Certificate{Fingerprint: fingerprint, NotAfter: time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)}}
}
func checkAlert(t *testing.T, s *Store, e Endpoint, status string, days int, fp string) Check {
	t.Helper()
	due(t, s, e.ID)
	c, err := s.Check(context.Background(), e.ID, func(context.Context, Endpoint) certificate.Result { return resultFor(e, status, days, fp) })
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestPostgresAlertLifecycle(t *testing.T) {
	s := &Store{Pool: testutil.Database(t)}
	ctx := context.Background()
	e := newEndpoint(t, s, "example.com")
	checkAlert(t, s, e, "healthy", 90, "a")
	page, err := s.Alerts(ctx, "all", "", 20, 0)
	if err != nil || len(page.Items) != 0 {
		t.Fatal(page, err)
	}
	checkAlert(t, s, e, "warning", 20, "a")
	checkAlert(t, s, e, "warning", 19, "a")
	page, err = s.Alerts(ctx, "active", "", 20, 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].Threshold != 30 {
		t.Fatal(page, err)
	}
	checkAlert(t, s, e, "critical", 5, "a")
	checkAlert(t, s, e, "critical", 3, "a")
	checkAlert(t, s, e, "critical", 2, "b")
	checkAlert(t, s, e, "healthy", 90, "c")
	checkAlert(t, s, e, "healthy", 89, "c")
	page, err = s.Alerts(ctx, "all", e.ID, 2, 0)
	if err != nil || len(page.Items) != 2 || page.Summary != (AlertSummary{Resolved: 4, Info: 1}) {
		t.Fatal(page, err)
	}
	checkAlert(t, s, e, "warning", 20, "c")
	if _, err = s.Update(ctx, e.ID, func(i *Input) error { i.Enabled = false; return nil }); err != nil {
		t.Fatal(err)
	}
	page, err = s.Alerts(ctx, "active", e.ID, 20, 0)
	if err != nil || page.Summary.Active != 1 {
		t.Fatal("pausa indicou recuperação", page, err)
	}
	e, err = s.Update(ctx, e.ID, func(i *Input) error { i.Hostname = "new.example.com"; i.Enabled = true; return nil })
	if err != nil {
		t.Fatal(err)
	}
	page, err = s.Alerts(ctx, "active", e.ID, 20, 0)
	if err != nil || len(page.Items) != 0 || page.Summary.Info != 1 {
		t.Fatal(page, err)
	}
	due(t, s, e.ID)
	if _, err = s.Check(ctx, e.ID, func(context.Context, Endpoint) certificate.Result {
		r := resultFor(e, "error", 0, "")
		r.Code = "dns_error"
		r.Certificate = nil
		return r
	}); err != nil {
		t.Fatal(err)
	}
	if err = s.Archive(ctx, e.ID); err != nil {
		t.Fatal(err)
	}
	page, err = s.Alerts(ctx, "resolved", e.ID, 20, 0)
	if err != nil || len(page.Items) != 7 || page.Summary != (AlertSummary{Resolved: 6, Info: 1}) {
		t.Fatal(page, err)
	}
	seen := map[string]bool{}
	for _, a := range page.Items {
		if a.Resolution != nil {
			seen[*a.Resolution] = true
		}
	}
	for _, reason := range []string{"changed", "recovered", "target_changed", "archived"} {
		if !seen[reason] {
			t.Fatal("motivo ausente", reason)
		}
	}
	if _, err = s.Alerts(ctx, "invalid", "", 20, 0); err != ErrInvalid {
		t.Fatal(err)
	}
	empty, err := s.Alerts(ctx, "all", "00000000-0000-0000-0000-000000000000", 20, 0)
	if err != nil || len(empty.Items) != 0 {
		t.Fatal(empty, err)
	}
}
func TestPostgresAlertAtomicityAndRetention(t *testing.T) {
	s := &Store{Pool: testutil.Database(t)}
	ctx := context.Background()
	e := newEndpoint(t, s, "example.com")
	first := checkAlert(t, s, e, "warning", 20, "a")
	if _, err := s.Pool.Exec(ctx, "ALTER TABLE alert_events ADD CONSTRAINT reject_test CHECK(event_type<>'dns_error')"); err != nil {
		t.Fatal(err)
	}
	due(t, s, e.ID)
	before, err := s.Get(ctx, e.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Check(ctx, e.ID, func(context.Context, Endpoint) certificate.Result {
		r := resultFor(e, "error", 0, "")
		r.Code = "dns_error"
		return r
	})
	if err == nil {
		t.Fatal("evento rejeitado não reverteu check")
	}
	after, err := s.Get(ctx, e.ID)
	if err != nil || after.LastCheck != nil || !after.NextCheck.Equal(*before.NextCheck) {
		t.Fatal(after, err)
	}
	page, err := s.Alerts(ctx, "all", e.ID, 20, 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].ResolvedAt != nil {
		t.Fatal("alerta anterior alterado sem commit", page, err)
	}
	second := checkAlert(t, s, e, "warning", 19, "a")
	if _, err = s.Pool.Exec(ctx, "UPDATE certificate_checks SET checked_at=now()-interval '200 days' WHERE id=$1", first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Pool.Exec(ctx, "UPDATE certificate_checks SET checked_at=now()-interval '190 days' WHERE id=$1", second.ID); err != nil {
		t.Fatal(err)
	}
	if n, err := s.PruneHistory(ctx, 180, 100); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	page, err = s.Alerts(ctx, "active", e.ID, 20, 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].CheckID != nil || page.Items[0].Snapshot.Certificate == nil {
		t.Fatal("snapshot perdido pela retenção", page, err)
	}
}
