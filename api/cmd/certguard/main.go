package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/httpapi"
	"github.com/franklinlem/certguard/api/internal/scheduler"
	"github.com/franklinlem/certguard/api/internal/storage"
	"github.com/franklinlem/certguard/api/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	if e := run(); e != nil {
		slog.Error("operação não concluída", "erro", e.Error())
		os.Exit(1)
	}
}
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	args := os.Args[1:]
	mode := "serve"
	if len(args) > 0 {
		mode = args[0]
	}
	if mode == "check" {
		if len(args) != 2 {
			return fmt.Errorf("uso: certguard check hostname")
		}
		c, _ := certificate.New(10*time.Second, 30, 7)
		result := c.Check(ctx, args[1], 443)
		if e := json.NewEncoder(os.Stdout).Encode(result); e != nil {
			return e
		}
		if result.Code != "success" {
			return fmt.Errorf("verificação: %s", result.Code)
		}
		return nil
	}
	if mode == "healthcheck" {
		client := http.Client{Timeout: 2 * time.Second}
		res, e := client.Get("http://127.0.0.1:8080/health")
		if e != nil {
			return fmt.Errorf("API indisponível")
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return fmt.Errorf("API indisponível")
		}
		return nil
	}
	if mode != "serve" && mode != "migrate" && mode != "admin-create" {
		return fmt.Errorf("comando desconhecido")
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return fmt.Errorf("DATABASE_URL obrigatório")
	}
	pool, e := pgxpool.New(ctx, url)
	if e != nil {
		return fmt.Errorf("configuração PostgreSQL inválida")
	}
	defer pool.Close()
	if mode == "migrate" {
		mctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if e = migrations.Up(mctx, pool); e != nil {
			return fmt.Errorf("falha ao aplicar migrações; verifique acesso e schema do banco")
		}
		slog.Info("migrações aplicadas")
		return nil
	}
	if mode == "admin-create" {
		if len(args) != 2 {
			return fmt.Errorf("uso: certguard admin-create email (senha via stdin)")
		}
		password, err := io.ReadAll(io.LimitReader(os.Stdin, 74))
		if err != nil {
			return fmt.Errorf("não foi possível ler a senha")
		}
		secret := strings.TrimSuffix(strings.TrimSuffix(string(password), "\n"), "\r")
		if len(password) > 73 {
			return fmt.Errorf("senha excede limite")
		}
		if err = (&storage.Store{Pool: pool}).Bootstrap(ctx, args[1], secret); err != nil {
			return fmt.Errorf("bootstrap não concluído; confira senha, e-mail, banco e existência de administrador")
		}
		slog.Info("administrador inicial criado")
		return nil
	}
	cfg, err := scheduler.Load(os.Getenv)
	if err != nil {
		return err
	}
	app, err := httpapi.NewApp(&storage.Store{Pool: pool}, os.Getenv("PUBLIC_APP_URL"), os.Getenv("APP_ENV") == "development")
	if err != nil {
		return err
	}
	app.Background = cfg
	if cfg.Enabled || cfg.RetentionEnabled {
		pc, err := pgxpool.ParseConfig(url)
		if err != nil {
			return fmt.Errorf("configuração PostgreSQL inválida")
		}
		pc.MaxConns = int32(cfg.Workers + 1)
		backgroundPool, err := pgxpool.NewWithConfig(ctx, pc)
		if err != nil {
			return fmt.Errorf("pool de agendamento indisponível")
		}
		defer backgroundPool.Close()
		backgroundCtx, cancel := context.WithCancel(ctx)
		finished := make(chan struct{})
		go func() {
			defer close(finished)
			_ = scheduler.Run(backgroundCtx, &storage.Store{Pool: backgroundPool}, cfg)
		}()
		defer func() { cancel(); <-finished }()
	}
	slog.Info("configuração de tarefas", "agendamento", cfg.Enabled, "workers", cfg.Workers, "retenção", cfg.RetentionEnabled)
	server := &http.Server{Addr: ":8080", Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16384}
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	slog.Info("API iniciada", "porta", 8080)
	select {
	case e = <-done:
		if e != http.ErrServerClosed {
			return e
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
	return nil
}
