package storage

import (
	"context"
	"github.com/franklinlem/certguard/api/internal/testutil"
	"sync"
	"testing"
	"time"
)

func loadFixture(tb testing.TB) *Store {
	tb.Helper()
	s := &Store{Pool: testutil.Database(tb)}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := s.Pool.Exec(ctx, `INSERT INTO monitored_endpoints(name,hostname) SELECT 'Carga '||i,'load-'||i||'.example.com' FROM generate_series(1,1000) i;
 INSERT INTO certificate_checks(endpoint_id,checked_at,status,code,duration_ms,metadata)
 SELECT e.id,now()-n*interval '1 minute','error','dns_error',1,jsonb_build_object('hostname',e.hostname,'port',443,'checked_at',now()-n*interval '1 minute','status','error','code','dns_error') FROM monitored_endpoints e CROSS JOIN generate_series(1,3) n;
 ANALYZE monitored_endpoints; ANALYZE certificate_checks;`); err != nil {
		tb.Fatal(err)
	}
	return s
}
func TestPostgresConcurrentReadSmoke(t *testing.T) {
	s := loadFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			for range 10 {
				d, err := s.Dashboard(ctx, 20, 0)
				if err != nil || d.Summary.Total != 1000 || d.Summary.Errors != 1000 || len(d.Items) != 20 {
					t.Error(d.Summary, err)
					return
				}
				m, err := s.Metrics(ctx)
				if err != nil || m.Endpoints != 1000 {
					t.Error(m, err)
					return
				}
			}
		})
	}
	workers.Wait()
}
func BenchmarkPostgresDashboard(b *testing.B) {
	s := loadFixture(b)
	b.ResetTimer()
	for range b.N {
		if _, err := s.Dashboard(context.Background(), 20, 0); err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkPostgresMetrics(b *testing.B) {
	s := loadFixture(b)
	b.ResetTimer()
	for range b.N {
		if _, err := s.Metrics(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}
