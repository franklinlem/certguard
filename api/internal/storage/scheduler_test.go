package storage

import (
	"context"
	"errors"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/testutil"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newEndpoint(t *testing.T, s *Store, host string) Endpoint {
	t.Helper()
	i := DefaultInput()
	i.Name = host
	i.Hostname = host
	i.Interval = 5
	e, err := s.Create(context.Background(), i)
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func due(t *testing.T, s *Store, id string) {
	t.Helper()
	if _, err := s.Pool.Exec(context.Background(), "UPDATE monitored_endpoints SET next_check_at=now()-interval '1 minute',last_check_at=NULL WHERE id=$1", id); err != nil {
		t.Fatal(err)
	}
}
func offline(_ context.Context, e Endpoint) certificate.Result {
	return certificate.Result{Hostname: e.Hostname, Port: e.Port, CheckedAt: time.Now().UTC(), Status: "error", Code: "blocked_target"}
}

func TestPostgresScheduleLifecycle(t *testing.T) {
	s := &Store{Pool: testutil.Database(t)}
	ctx := context.Background()
	e := newEndpoint(t, s, "example.com")
	if e.NextCheck == nil || time.Until(*e.NextCheck) > 31*time.Second || time.Until(*e.NextCheck) < -time.Second {
		t.Fatal("agenda inicial", e)
	}
	original := *e.NextCheck
	e, err := s.Update(ctx, e.ID, func(i *Input) error { i.Name = "Outro nome"; return nil })
	if err != nil || e.NextCheck == nil || !e.NextCheck.Equal(original) {
		t.Fatal("renomear alterou agenda", err)
	}
	due(t, s, e.ID)
	if worked, err := s.CheckDue(ctx, offline); err != nil || !worked {
		t.Fatal(worked, err)
	}
	e, err = s.Get(ctx, e.ID)
	if err != nil || e.LastCheck == nil || e.NextCheck == nil {
		t.Fatal(e, err)
	}
	delta := e.NextCheck.Sub(*e.LastCheck)
	if delta < 5*time.Minute || delta > 331*time.Second {
		t.Fatal("intervalo/jitter", delta)
	}
	if worked, err := s.CheckDue(ctx, offline); err != nil || worked {
		t.Fatal("endpoint futuro selecionado", err)
	}
	if _, err = s.Check(ctx, e.ID, offline); !errors.Is(err, ErrBusy) {
		t.Fatal("cooldown manual", err)
	}
	e, err = s.Update(ctx, e.ID, func(i *Input) error { i.Enabled = false; return nil })
	if err != nil || e.NextCheck != nil {
		t.Fatal("pausa", err)
	}
	due(t, s, e.ID) // Mesmo uma agenda residual não habilita um endpoint pausado.
	if worked, err := s.CheckDue(ctx, offline); err != nil || worked {
		t.Fatal("endpoint pausado selecionado", err)
	}
	e, err = s.Update(ctx, e.ID, func(i *Input) error { i.Enabled = true; return nil })
	if err != nil || e.NextCheck == nil || time.Until(*e.NextCheck) > 31*time.Second {
		t.Fatal("retomada", err)
	}
	due(t, s, e.ID)
	if _, err = s.Check(ctx, e.ID, offline); err != nil {
		t.Fatal(err)
	}
	if worked, err := s.CheckDue(ctx, offline); err != nil || worked {
		t.Fatal("manual não reagendou", err)
	}
	if err = s.Archive(ctx, e.ID); err != nil {
		t.Fatal(err)
	}
	due(t, s, e.ID)
	if worked, err := s.CheckDue(ctx, offline); err != nil || worked {
		t.Fatal("arquivado selecionado", err)
	}
}

func TestPostgresScheduledConcurrencyAndCancellation(t *testing.T) {
	s := &Store{Pool: testutil.Database(t)}
	ctx := context.Background()
	e := newEndpoint(t, s, "example.com")
	due(t, s, e.ID)
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := s.CheckDue(ctx, func(ctx context.Context, e Endpoint) certificate.Result {
			close(started)
			<-release
			return offline(ctx, e)
		})
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("check não iniciou")
	}
	worked, err := s.CheckDue(ctx, offline)
	if err != nil || worked {
		t.Error("lock não ignorado", err)
	}
	if _, err = s.Check(ctx, e.ID, offline); !Conflict(err) {
		t.Error("manual sobrepôs check", err)
	}
	close(release)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	due(t, s, e.ID)
	cancelled, cancel := context.WithCancel(ctx)
	_, err = s.CheckDue(cancelled, func(ctx context.Context, e Endpoint) certificate.Result { cancel(); return offline(ctx, e) })
	if !errors.Is(err, context.Canceled) {
		t.Fatal("cancelamento não propagado", err)
	}
	h, err := s.History(ctx, e.ID, 10, 0)
	if err != nil || len(h) != 1 {
		t.Fatal("cancelamento persistiu resultado", h, err)
	}
	var count atomic.Int32
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			_, err := s.CheckDue(ctx, func(ctx context.Context, e Endpoint) certificate.Result { count.Add(1); return offline(ctx, e) })
			if err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if count.Load() != 1 {
		t.Fatal("execução duplicada", count.Load())
	}
	page, err := s.Alerts(ctx, "all", e.ID, 20, 0)
	if err != nil || len(page.Items) != 1 {
		t.Fatal("evento duplicado após disputa de workers", page, err)
	}
}

func TestPostgresRetention(t *testing.T) {
	s := &Store{Pool: testutil.Database(t)}
	ctx := context.Background()
	a := newEndpoint(t, s, "example.com")
	b := newEndpoint(t, s, "example.org")
	for _, item := range []struct {
		id  string
		age int
	}{{a.ID, 200}, {a.ID, 190}, {a.ID, 1}, {b.ID, 300}, {b.ID, 300}} {
		if _, err := s.Pool.Exec(ctx, `INSERT INTO certificate_checks(endpoint_id,checked_at,status,code,duration_ms,metadata) VALUES($1,now()-$2*interval '1 day','error','blocked_target',0,'{}')`, item.id, item.age); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Archive(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	// Um lote remove no máximo um; lotes seguintes preservam o último mesmo se antigo.
	if n, err := s.PruneHistory(ctx, 180, 1); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	if n, err := s.PruneHistory(ctx, 180, 100); err != nil || n != 2 {
		t.Fatal(n, err)
	}
	if n, err := s.PruneHistory(ctx, 180, 100); err != nil || n != 0 {
		t.Fatal(n, err)
	}
	for _, id := range []string{a.ID, b.ID} {
		var n int
		if err := s.Pool.QueryRow(ctx, "SELECT count(*) FROM certificate_checks WHERE endpoint_id=$1", id).Scan(&n); err != nil || n != 1 {
			t.Fatal(n, err)
		}
	}
	for _, args := range [][2]int{{29, 1}, {180, 0}, {180, 1001}, {36501, 1}} {
		if _, err := s.PruneHistory(ctx, args[0], args[1]); !errors.Is(err, ErrInvalid) {
			t.Fatal("limites ignorados", err)
		}
	}
}
