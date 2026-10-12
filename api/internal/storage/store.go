package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/franklinlem/certguard/api/internal/auth"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/security"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

var ErrInvalid = errors.New("dados inválidos")
var ErrBusy = errors.New("verificação em andamento ou intervalo mínimo não atingido")
var ErrDisabled = errors.New("endpoint desativado")

type Store struct{ Pool *pgxpool.Pool }
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
type Endpoint struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Hostname  string     `json:"hostname"`
	Port      int        `json:"port"`
	Enabled   bool       `json:"enabled"`
	Interval  int        `json:"interval_minutes"`
	Warning   int        `json:"warning_days"`
	Critical  int        `json:"critical_days"`
	LastCheck *time.Time `json:"last_check_at"`
}
type Input struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Enabled  bool   `json:"enabled"`
	Interval int    `json:"interval_minutes"`
	Warning  int    `json:"warning_days"`
	Critical int    `json:"critical_days"`
}

func DefaultInput() Input {
	return Input{Port: 443, Enabled: true, Interval: 360, Warning: 30, Critical: 7}
}
func (i *Input) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	h, e := security.Normalize(i.Hostname)
	if e != nil || len(i.Name) < 1 || len(i.Name) > 200 || i.Port != 443 || i.Interval < 5 || i.Interval > 10080 || i.Critical < 0 || i.Warning <= i.Critical || i.Warning > 36500 {
		return ErrInvalid
	}
	i.Hostname = h
	return nil
}
func (e Endpoint) Input() Input {
	return Input{e.Name, e.Hostname, e.Port, e.Enabled, e.Interval, e.Warning, e.Critical}
}

const columns = "id::text,name,hostname,port,enabled,interval_minutes,warning_days,critical_days,last_check_at"

func scan(row pgx.Row) (Endpoint, error) {
	var e Endpoint
	err := row.Scan(&e.ID, &e.Name, &e.Hostname, &e.Port, &e.Enabled, &e.Interval, &e.Warning, &e.Critical, &e.LastCheck)
	return e, err
}
func (s *Store) Bootstrap(ctx context.Context, email, password string) error {
	email, e := auth.Email(email)
	if e != nil {
		return e
	}
	hash, e := auth.HashPassword(password)
	if e != nil {
		return e
	}
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(724319002)"); e != nil {
		return e
	}
	var count int
	if e = tx.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&count); e != nil {
		return e
	}
	if count != 0 {
		return errors.New("administrador já inicializado")
	}
	if _, e = tx.Exec(ctx, "INSERT INTO users(email,password_hash) VALUES($1,$2)", email, hash); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
func (s *Store) Credentials(ctx context.Context, email string) (User, string, error) {
	var u User
	var h string
	e := s.Pool.QueryRow(ctx, "SELECT id::text,email,password_hash FROM users WHERE email=$1", email).Scan(&u.ID, &u.Email, &h)
	return u, h, e
}
func (s *Store) NewSession(ctx context.Context, userID, hash string) error {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	// Serializa logins do mesmo usuário e mantém no máximo dez sessões ativas.
	if _, e = tx.Exec(ctx, "SELECT id FROM users WHERE id=$1 FOR UPDATE", userID); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "DELETE FROM sessions WHERE expires_at<=now()"); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "DELETE FROM sessions WHERE token_hash IN (SELECT token_hash FROM sessions WHERE user_id=$1 ORDER BY created_at DESC OFFSET 9)", userID); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO sessions(token_hash,user_id,expires_at) VALUES($1,$2,now()+interval '8 hours')", hash, userID); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
func (s *Store) Session(ctx context.Context, hash string) (User, error) {
	var u User
	e := s.Pool.QueryRow(ctx, "SELECT u.id::text,u.email FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.expires_at>now()", hash).Scan(&u.ID, &u.Email)
	return u, e
}
func (s *Store) Logout(ctx context.Context, hash string) error {
	_, e := s.Pool.Exec(ctx, "DELETE FROM sessions WHERE token_hash=$1", hash)
	return e
}
func (s *Store) List(ctx context.Context, limit, offset int) ([]Endpoint, error) {
	rows, e := s.Pool.Query(ctx, "SELECT "+columns+" FROM monitored_endpoints WHERE deleted_at IS NULL ORDER BY created_at DESC,id LIMIT $1 OFFSET $2", limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Endpoint{}
	for rows.Next() {
		v, e := scan(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Get(ctx context.Context, id string) (Endpoint, error) {
	return scan(s.Pool.QueryRow(ctx, "SELECT "+columns+" FROM monitored_endpoints WHERE id=$1 AND deleted_at IS NULL", id))
}
func (s *Store) Create(ctx context.Context, i Input) (Endpoint, error) {
	if e := i.Validate(); e != nil {
		return Endpoint{}, e
	}
	return scan(s.Pool.QueryRow(ctx, "INSERT INTO monitored_endpoints(name,hostname,port,enabled,interval_minutes,warning_days,critical_days) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING "+columns, i.Name, i.Hostname, i.Port, i.Enabled, i.Interval, i.Warning, i.Critical))
}

// Update aplica patch sob lock para não perder alterações simultâneas.
func (s *Store) Update(ctx context.Context, id string, patch func(*Input) error) (Endpoint, error) {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return Endpoint{}, e
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	v, e := scan(tx.QueryRow(ctx, "SELECT "+columns+" FROM monitored_endpoints WHERE id=$1 AND deleted_at IS NULL FOR UPDATE NOWAIT", id))
	if e != nil {
		return Endpoint{}, e
	}
	i := v.Input()
	if e = patch(&i); e != nil {
		return Endpoint{}, e
	}
	if e = i.Validate(); e != nil {
		return Endpoint{}, e
	}
	v, e = scan(tx.QueryRow(ctx, "UPDATE monitored_endpoints SET name=$2,hostname=$3,port=$4,enabled=$5,interval_minutes=$6,warning_days=$7,critical_days=$8,updated_at=now() WHERE id=$1 RETURNING "+columns, id, i.Name, i.Hostname, i.Port, i.Enabled, i.Interval, i.Warning, i.Critical))
	if e != nil {
		return Endpoint{}, e
	}
	return v, tx.Commit(ctx)
}
func (s *Store) Archive(ctx context.Context, id string) error {
	tag, e := s.Pool.Exec(ctx, "UPDATE monitored_endpoints SET deleted_at=now(),enabled=false,updated_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if e == nil && tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return e
}

type Check struct {
	ID         string             `json:"id"`
	EndpointID string             `json:"endpoint_id"`
	Result     certificate.Result `json:"result"`
}
type Inspect func(context.Context, Endpoint) certificate.Result

func (s *Store) Check(ctx context.Context, id string, inspect Inspect) (Check, error) {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return Check{}, e
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	endpoint, e := scan(tx.QueryRow(ctx, "SELECT "+columns+" FROM monitored_endpoints WHERE id=$1 AND deleted_at IS NULL FOR UPDATE NOWAIT", id))
	if e != nil {
		return Check{}, e
	}
	if !endpoint.Enabled {
		return Check{}, ErrDisabled
	}
	if endpoint.LastCheck != nil && time.Since(*endpoint.LastCheck) < 30*time.Second {
		return Check{}, ErrBusy
	}
	result := inspect(ctx, endpoint)
	data, e := json.Marshal(result)
	if e != nil {
		return Check{}, e
	}
	out := Check{EndpointID: id, Result: result}
	var ip *string
	if result.ResolvedIP != "" {
		ip = &result.ResolvedIP
	}
	if e = tx.QueryRow(ctx, "INSERT INTO certificate_checks(endpoint_id,checked_at,status,code,duration_ms,resolved_ip,metadata) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id::text", id, result.CheckedAt, result.Status, result.Code, result.DurationMS, ip, data).Scan(&out.ID); e != nil {
		return Check{}, e
	}
	if _, e = tx.Exec(ctx, "UPDATE monitored_endpoints SET last_check_at=now(),updated_at=now() WHERE id=$1", id); e != nil {
		return Check{}, e
	}
	return out, tx.Commit(ctx)
}
func (s *Store) History(ctx context.Context, id string, limit, offset int) ([]Check, error) {
	// Histórico permanece consultável após arquivamento.
	var exists bool
	if e := s.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM monitored_endpoints WHERE id=$1)", id).Scan(&exists); e != nil {
		return nil, e
	}
	if !exists {
		return nil, pgx.ErrNoRows
	}
	rows, e := s.Pool.Query(ctx, "SELECT id::text,metadata FROM certificate_checks WHERE endpoint_id=$1 ORDER BY checked_at DESC,id LIMIT $2 OFFSET $3", id, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Check{}
	for rows.Next() {
		c := Check{EndpointID: id}
		var raw []byte
		if e = rows.Scan(&c.ID, &raw); e != nil {
			return nil, e
		}
		if e = json.Unmarshal(raw, &c.Result); e != nil {
			return nil, e
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func Conflict(e error) bool {
	var p *pgconn.PgError
	return errors.As(e, &p) && (p.Code == "23505" || p.Code == "55P03")
}
