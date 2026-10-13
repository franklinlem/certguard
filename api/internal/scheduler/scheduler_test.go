package scheduler

import (
	"context"
	"errors"
	"github.com/franklinlem/certguard/api/internal/storage"
	"sync/atomic"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	c, err := Load(func(string) string { return "" })
	if err != nil || c.Enabled || c.RetentionEnabled || c.Workers != 10 || c.RetentionDays != 180 {
		t.Fatal(c, err)
	}
	for key, values := range map[string][]string{
		"SCHEDULER_ENABLED": {"maybe"}, "RETENTION_ENABLED": {"maybe"},
		"CHECK_WORKERS": {"0", "26", "abc"}, "CHECK_TIMEOUT": {"0s", "31s", "abc"},
		"RETENTION_DAYS": {"29", "36501"}, "RETENTION_BATCH": {"0", "1001"},
	} {
		for _, v := range values {
			t.Run(key+v, func(t *testing.T) {
				if _, err := Load(func(k string) string {
					if k == key {
						return v
					}
					return ""
				}); err == nil {
					t.Fatal("configuração inválida aceita")
				}
			})
		}
	}
	c, err = Load(func(k string) string {
		switch k {
		case "SCHEDULER_ENABLED", "RETENTION_ENABLED":
			return "true"
		case "CHECK_WORKERS":
			return "25"
		case "CHECK_TIMEOUT":
			return "30s"
		}
		return ""
	})
	if err != nil || !c.Enabled || !c.RetentionEnabled || c.Workers != 25 || c.Timeout != 30*time.Second {
		t.Fatal(c, err)
	}
	if Run(context.Background(), nil, Config{}) == nil {
		t.Fatal("configuração inválida executada")
	}
}

type fakeStore struct {
	check func(context.Context) (bool, error)
	prune func(context.Context, int, int) (int64, error)
}

func (f fakeStore) CheckDue(ctx context.Context, _ storage.Inspect) (bool, error) {
	return f.check(ctx)
}
func (f fakeStore) PruneHistory(ctx context.Context, d, b int) (int64, error) {
	return f.prune(ctx, d, b)
}

func TestWorkersBoundedAndCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, _ := Load(func(string) string { return "" })
	c.Enabled = true
	c.Workers = 3
	entered := make(chan struct{}, 10)
	done := make(chan struct{})
	var active atomic.Int32
	store := fakeStore{check: func(ctx context.Context) (bool, error) {
		active.Add(1)
		defer active.Add(-1)
		entered <- struct{}{}
		<-ctx.Done()
		return false, ctx.Err()
	}}
	go func() { run(ctx, store, c, nil, wait, func() time.Duration { return 0 }); close(done) }()
	for range 3 {
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("workers não iniciaram")
		}
	}
	if active.Load() != 3 {
		t.Fatal(active.Load())
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("encerramento bloqueado")
	}
	if active.Load() != 0 {
		t.Fatal("worker pendente")
	}
}

func TestBackoffAndRetention(t *testing.T) {
	c, _ := Load(func(string) string { return "" })
	c.Enabled = true
	c.Workers = 1
	for _, result := range []struct {
		worked bool
		err    error
		delay  time.Duration
	}{{false, nil, 5 * time.Second}, {false, errors.New("db"), 5 * time.Second}, {true, nil, 250 * time.Millisecond}} {
		calls := 0
		run(context.Background(), fakeStore{check: func(context.Context) (bool, error) { return result.worked, result.err }}, c, nil, func(_ context.Context, d time.Duration) bool {
			calls++
			if calls == 1 {
				if d != 7*time.Second {
					t.Error("jitter ignorado")
				}
				return true
			}
			if d != result.delay {
				t.Error(d)
			}
			return false
		}, func() time.Duration { return 7 * time.Second })
		if calls != 2 {
			t.Fatal(calls)
		}
	}
	c.Enabled = false
	c.RetentionEnabled = true
	calls := 0
	pruned := false
	run(context.Background(), fakeStore{prune: func(_ context.Context, d, b int) (int64, error) {
		pruned = true
		if d != 180 || b != 500 {
			t.Error(d, b)
		}
		return 1, nil
	}}, c, nil, func(_ context.Context, d time.Duration) bool {
		if d != time.Minute {
			t.Error(d)
		}
		calls++
		return calls == 1
	}, func() time.Duration { return 0 })
	if !pruned || calls != 2 {
		t.Fatal("retenção não executada")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if wait(ctx, time.Hour) {
		t.Fatal("espera ignorou cancelamento")
	}
}
