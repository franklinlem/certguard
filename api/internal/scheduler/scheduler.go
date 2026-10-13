// Package scheduler executa verificações e retenção com concorrência limitada.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"sync"
	"time"

	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/storage"
)

type Config struct {
	Enabled          bool          `json:"scheduler_enabled"`
	Workers          int           `json:"check_workers"`
	Timeout          time.Duration `json:"-"`
	RetentionEnabled bool          `json:"retention_enabled"`
	RetentionDays    int           `json:"retention_days"`
	RetentionBatch   int           `json:"-"`
}

func Load(get func(string) string) (Config, error) {
	c := Config{Workers: 10, Timeout: 10 * time.Second, RetentionDays: 180, RetentionBatch: 500}
	for key, dst := range map[string]*bool{"SCHEDULER_ENABLED": &c.Enabled, "RETENTION_ENABLED": &c.RetentionEnabled} {
		if v := get(key); v != "" {
			b, e := strconv.ParseBool(v)
			if e != nil {
				return c, fmt.Errorf("%s inválido", key)
			}
			*dst = b
		}
	}
	for key, dst := range map[string]*int{"CHECK_WORKERS": &c.Workers, "RETENTION_DAYS": &c.RetentionDays, "RETENTION_BATCH": &c.RetentionBatch} {
		if v := get(key); v != "" {
			n, e := strconv.Atoi(v)
			if e != nil {
				return c, fmt.Errorf("%s inválido", key)
			}
			*dst = n
		}
	}
	if v := get("CHECK_TIMEOUT"); v != "" {
		d, e := time.ParseDuration(v)
		if e != nil {
			return c, fmt.Errorf("CHECK_TIMEOUT inválido")
		}
		c.Timeout = d
	}
	return c, c.Validate()
}

func (c Config) Validate() error {
	if c.Workers < 1 || c.Workers > 25 || c.Timeout < time.Second || c.Timeout > 30*time.Second || c.RetentionDays < 30 || c.RetentionDays > 36500 || c.RetentionBatch < 1 || c.RetentionBatch > 1000 {
		return fmt.Errorf("configuração de agendamento ou retenção fora dos limites")
	}
	return nil
}

type backend interface {
	CheckDue(context.Context, storage.Inspect) (bool, error)
	PruneHistory(context.Context, int, int) (int64, error)
}

// Run aguarda o encerramento de todos os workers após o cancelamento.
func Run(ctx context.Context, store backend, c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	inspect := func(ctx context.Context, e storage.Endpoint) certificate.Result {
		checker, _ := certificate.New(c.Timeout, e.Warning, e.Critical)
		return checker.Check(ctx, e.Hostname, e.Port)
	}
	run(ctx, store, c, inspect, wait, func() time.Duration { return time.Duration(rand.Int64N(int64(30 * time.Second))) })
	return nil
}

type pause func(context.Context, time.Duration) bool

func wait(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return ctx.Err() == nil
	}
}

func run(ctx context.Context, store backend, c Config, inspect storage.Inspect, sleep pause, jitter func() time.Duration) {
	var wg sync.WaitGroup
	if c.Enabled {
		for range c.Workers {
			wg.Go(func() {
				if !sleep(ctx, jitter()) {
					return
				}
				for ctx.Err() == nil {
					op, cancel := context.WithTimeout(ctx, c.Timeout+5*time.Second)
					worked, err := store.CheckDue(op, inspect)
					cancel()
					delay := 250 * time.Millisecond
					if !worked {
						delay = 5 * time.Second
					}
					if err != nil && ctx.Err() == nil {
						slog.Error("agendamento: falha de persistência; nova tentativa em 5s")
						delay = 5 * time.Second
					}
					if !sleep(ctx, delay) {
						return
					}
				}
			})
		}
	}
	if c.RetentionEnabled {
		wg.Go(func() {
			for sleep(ctx, time.Minute) {
				op, cancel := context.WithTimeout(ctx, 10*time.Second)
				count, err := store.PruneHistory(op, c.RetentionDays, c.RetentionBatch)
				cancel()
				if err != nil && ctx.Err() == nil {
					slog.Error("retenção: lote não concluído")
				} else if count > 0 {
					slog.Info("retenção: resultados antigos removidos", "quantidade", count)
				}
			}
		})
	}
	wg.Wait()
}
