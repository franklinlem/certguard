package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/testutil"
	"testing"
	"time"
)

func TestPostgresDashboard(t *testing.T) {
	s := &Store{Pool: testutil.Database(t)}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	empty, err := s.Dashboard(ctx, 20, 0)
	if err != nil || empty.Summary.Total != 0 || len(empty.Items) != 0 || empty.Items == nil {
		t.Fatal(empty, err)
	}
	now := time.Now().UTC()
	serial := 0
	insert := func(e Endpoint, status string, days *int, at time.Time) {
		t.Helper()
		serial++
		r := certificate.Result{Hostname: e.Hostname, Port: e.Port, CheckedAt: at, Status: status, Code: "success", DaysRemaining: days}
		if days != nil {
			r.Certificate = &certificate.Certificate{NotAfter: now.Add(time.Duration(*days) * 24 * time.Hour)}
		}
		b, _ := json.Marshal(r)
		if _, err := s.Pool.Exec(ctx, `INSERT INTO certificate_checks(id,endpoint_id,checked_at,status,code,duration_ms,metadata) VALUES($1::uuid,$2,$3,$4,'success',0,$5)`, fmt.Sprintf("00000000-0000-0000-0000-%012d", serial), e.ID, at, status, b); err != nil {
			t.Fatal(err)
		}
	}
	days45, days10, expired := 45, 10, -1
	a := newEndpoint(t, s, "a.example.com")
	insert(a, "healthy", &days45, now)
	b := newEndpoint(t, s, "b.example.com")
	insert(b, "warning", &days10, now)
	if _, err = s.Update(ctx, b.ID, func(i *Input) error { i.Enabled = false; return nil }); err != nil {
		t.Fatal(err)
	}
	c := newEndpoint(t, s, "c.example.com")
	insert(c, "critical", &expired, now)
	d := newEndpoint(t, s, "d.example.com")
	insert(d, "healthy", &days10, now.Add(-time.Hour))
	insert(d, "error", nil, now)
	newEndpoint(t, s, "e.example.com")
	f := newEndpoint(t, s, "f.example.com")
	insert(f, "healthy", &days10, now)
	if _, err = s.Update(ctx, f.ID, func(i *Input) error { i.Hostname = "changed.example.com"; return nil }); err != nil {
		t.Fatal(err)
	}
	g := newEndpoint(t, s, "g.example.com")
	insert(g, "critical", &days10, now)
	if err = s.Archive(ctx, g.ID); err != nil {
		t.Fatal(err)
	}
	h := newEndpoint(t, s, "h.example.com")
	insert(h, "healthy", &days45, now)
	insert(h, "critical", &days10, now)
	dashboard, err := s.Dashboard(ctx, 2, 1)
	want := Summary{Total: 7, Enabled: 6, Paused: 1, Healthy: 1, Warning: 1, Critical: 2, Errors: 1, Unchecked: 2, Expiring30: 2}
	if err != nil || dashboard.Summary != want || len(dashboard.Items) != 2 || dashboard.GeneratedAt.IsZero() {
		t.Fatal(dashboard, err)
	}
	all, err := s.Dashboard(ctx, 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range all.Items {
		if v.ID == f.ID && v.Latest != nil {
			t.Fatal("resultado de destino anterior reaproveitado")
		}
		if v.ID == h.ID && (v.Latest == nil || v.Latest.Status != "critical") {
			t.Fatal("desempate não determinístico")
		}
		if v.ID == d.ID && (v.Latest == nil || v.Latest.Certificate != nil) {
			t.Fatal("certificado antigo apresentado após falha")
		}
	}
	beyond, err := s.Dashboard(ctx, 2, 99)
	if err != nil || len(beyond.Items) != 0 || beyond.Summary != want {
		t.Fatal(beyond, err)
	}
	if _, err = s.Dashboard(ctx, 101, 0); err != ErrInvalid {
		t.Fatal(err)
	}
}
