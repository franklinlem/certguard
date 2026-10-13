package storage

import (
	"context"
	"errors"
	"github.com/franklinlem/certguard/api/internal/auth"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/testutil"
	"github.com/jackc/pgx/v5"
	"testing"
	"time"
)

func TestInput(t *testing.T) {
	for _, mutate := range []func(*Input){func(i *Input) { i.Hostname = "127.0.0.1" }, func(i *Input) { i.Hostname = "https://example.com" }, func(i *Input) { i.Name = " " }, func(i *Input) { i.Port = 80 }, func(i *Input) { i.Interval = 4 }, func(i *Input) { i.Warning = 7 }, func(i *Input) { i.Critical = -1 }, func(i *Input) { i.Warning = 36501 }} {
		i := DefaultInput()
		i.Name = "Teste"
		i.Hostname = "example.com"
		mutate(&i)
		if i.Validate() == nil {
			t.Fatal(i)
		}
	}
	i := DefaultInput()
	i.Name = " Exemplo "
	i.Hostname = "EXAMPLE.COM."
	if i.Validate() != nil || i.Hostname != "example.com" || i.Name != "Exemplo" {
		t.Fatal(i)
	}
}
func TestPostgresStorage(t *testing.T) {
	p := testutil.Database(t)
	s := &Store{Pool: p}
	ctx := context.Background()
	if e := s.Bootstrap(ctx, "admin@example.com", "Teste-local-12345"); e != nil {
		t.Fatal(e)
	}
	if e := s.Bootstrap(ctx, "other@example.com", "Teste-local-12345"); e == nil {
		t.Fatal("segundo bootstrap")
	}
	u, h, e := s.Credentials(ctx, "admin@example.com")
	if e != nil || !auth.Verify(h, "Teste-local-12345") {
		t.Fatal(e)
	}
	token := auth.Token()
	if e = s.NewSession(ctx, u.ID, auth.Hash(token)); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Session(ctx, token); !errors.Is(e, pgx.ErrNoRows) {
		t.Fatal("token em claro no banco")
	}
	if _, e = s.Session(ctx, auth.Hash(token)); e != nil {
		t.Fatal(e)
	}
	if _, e = p.Exec(ctx, "UPDATE sessions SET expires_at=now()-interval '1 second'"); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Session(ctx, auth.Hash(token)); !errors.Is(e, pgx.ErrNoRows) {
		t.Fatal("sessão expirada aceita")
	}
	i := DefaultInput()
	i.Name = "Exemplo"
	i.Hostname = "example.com"
	ep, e := s.Create(ctx, i)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Create(ctx, i); !Conflict(e) {
		t.Fatal("duplicata aceita", e)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	inspect := func(ctx context.Context, e Endpoint) certificate.Result {
		close(started)
		<-release
		return certificate.Result{Hostname: e.Hostname, Port: e.Port, CheckedAt: time.Now().UTC(), Status: "error", Code: "blocked_target"}
	}
	go func() { _, e := s.Check(ctx, ep.ID, inspect); done <- e }()
	<-started
	_, e = s.Check(ctx, ep.ID, func(context.Context, Endpoint) certificate.Result {
		t.Error("check concorrente executado")
		return certificate.Result{}
	})
	if !Conflict(e) {
		t.Error("lock não aplicado", e)
	}
	close(release)
	if e = <-done; e != nil {
		t.Fatal(e)
	}
	history, e := s.History(ctx, ep.ID, 10, 0)
	if e != nil || len(history) != 1 || history[0].Result.Code != "blocked_target" {
		t.Fatal(history, e)
	}
	_, e = s.Check(ctx, ep.ID, inspect)
	if !errors.Is(e, ErrBusy) {
		t.Fatal("cooldown ignorado", e)
	}
	ep, e = s.Update(ctx, ep.ID, func(i *Input) error { i.Enabled = false; return nil })
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Check(ctx, ep.ID, inspect); !errors.Is(e, ErrDisabled) {
		t.Fatal(e)
	}
	if e = s.Archive(ctx, ep.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Get(ctx, ep.ID); !errors.Is(e, pgx.ErrNoRows) {
		t.Fatal(e)
	}
	if h, e := s.History(ctx, ep.ID, 10, 0); e != nil || len(h) != 1 {
		t.Fatal("histórico apagado", e)
	}
	if _, e = s.Create(ctx, i); e != nil {
		t.Fatal("recadastro após arquivo", e)
	}
}
func TestPostgresCheckRollback(t *testing.T) {
	s := &Store{Pool: testutil.Database(t)}
	ctx := context.Background()
	i := DefaultInput()
	i.Name = "Teste"
	i.Hostname = "example.com"
	ep, e := s.Create(ctx, i)
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Check(ctx, ep.ID, func(context.Context, Endpoint) certificate.Result {
		return certificate.Result{Status: "invalid", CheckedAt: time.Now(), Code: "invalid"}
	})
	if e == nil {
		t.Fatal("resultado inválido salvo")
	}
	fresh, e := s.Get(ctx, ep.ID)
	if e != nil || fresh.LastCheck != nil {
		t.Fatal("timestamp fora da transação", e)
	}
	h, e := s.History(ctx, ep.ID, 10, 0)
	if e != nil || len(h) != 0 {
		t.Fatal(h, e)
	}
}
