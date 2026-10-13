package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/franklinlem/certguard/api/internal/alerts"
	"github.com/jackc/pgx/v5"
	"strings"
	"time"
)

var ErrUnavailable = errors.New("chave de canais indisponível")

type Channel struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}
type ChannelInput struct {
	Type   string               `json:"type"`
	Name   string               `json:"name"`
	Config alerts.ChannelConfig `json:"config"`
}

const channelCols = "id::text,type,name,enabled,created_at"

func channelScan(row pgx.Row) (Channel, error) {
	var c Channel
	err := row.Scan(&c.ID, &c.Type, &c.Name, &c.Enabled, &c.CreatedAt)
	return c, err
}
func (s *Store) Channels(ctx context.Context) ([]Channel, error) {
	rows, err := s.Pool.Query(ctx, "SELECT "+channelCols+" FROM alert_channels WHERE deleted_at IS NULL ORDER BY created_at,id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Channel{}
	for rows.Next() {
		c, err := channelScan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (s *Store) CreateChannel(ctx context.Context, i ChannelInput) (Channel, error) {
	i.Name = strings.TrimSpace(i.Name)
	if len(i.Name) < 1 || len(i.Name) > 200 || i.Config.Validate(i.Type) != nil {
		return Channel{}, ErrInvalid
	}
	if s.Cipher == nil {
		return Channel{}, ErrUnavailable
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Channel{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(724319003)"); err != nil {
		return Channel{}, err
	}
	var count int
	var id string
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM alert_channels WHERE deleted_at IS NULL").Scan(&count); err != nil {
		return Channel{}, err
	}
	if count >= 20 {
		return Channel{}, ErrBusy
	}
	if err = tx.QueryRow(ctx, "SELECT gen_random_uuid()::text").Scan(&id); err != nil {
		return Channel{}, err
	}
	sealed, err := s.Cipher.Seal(id, i.Type, i.Config)
	if err != nil {
		return Channel{}, ErrUnavailable
	}
	c, err := channelScan(tx.QueryRow(ctx, "INSERT INTO alert_channels(id,type,name,config_encrypted) VALUES($1,$2,$3,$4) RETURNING "+channelCols, id, i.Type, i.Name, sealed))
	if err != nil {
		return Channel{}, err
	}
	return c, tx.Commit(ctx)
}
func (s *Store) UpdateChannel(ctx context.Context, id string, name *string, enabled *bool, archive bool) (Channel, error) {
	if enabled != nil && *enabled && s.Cipher == nil {
		return Channel{}, ErrUnavailable
	}
	if name != nil {
		v := strings.TrimSpace(*name)
		name = &v
		if len(v) < 1 || len(v) > 200 {
			return Channel{}, ErrInvalid
		}
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Channel{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	c, err := channelScan(tx.QueryRow(ctx, "SELECT "+channelCols+" FROM alert_channels WHERE id=$1 AND deleted_at IS NULL FOR UPDATE NOWAIT", id))
	if err != nil {
		return Channel{}, err
	}
	if enabled != nil && *enabled {
		var sealed []byte
		if err = tx.QueryRow(ctx, "SELECT config_encrypted FROM alert_channels WHERE id=$1", id).Scan(&sealed); err != nil {
			return Channel{}, err
		}
		if _, err = s.Cipher.Open(id, c.Type, sealed); err != nil {
			return Channel{}, ErrUnavailable
		}
	}
	if name != nil {
		c.Name = *name
	}
	if enabled != nil {
		c.Enabled = *enabled
	}
	if archive {
		c.Enabled = false
	}
	c, err = channelScan(tx.QueryRow(ctx, "UPDATE alert_channels SET name=$2,enabled=$3,deleted_at=CASE WHEN $4 THEN clock_timestamp() ELSE deleted_at END,updated_at=clock_timestamp() WHERE id=$1 RETURNING "+channelCols, id, c.Name, c.Enabled, archive))
	if err != nil {
		return Channel{}, err
	}
	if !c.Enabled {
		code := "channel_paused"
		if archive {
			code = "channel_archived"
		}
		if _, err = tx.Exec(ctx, "UPDATE alert_deliveries SET status='cancelled',next_attempt_at=NULL,last_error=$2,updated_at=clock_timestamp() WHERE channel_id=$1 AND status='pending'", id, code); err != nil {
			return Channel{}, err
		}
	}
	return c, tx.Commit(ctx)
}

// VerifyChannels impede iniciar o executor com chave incompatível com canais ativos.
func (s *Store) VerifyChannels(ctx context.Context) error {
	rows, err := s.Pool.Query(ctx, "SELECT id::text,type,config_encrypted FROM alert_channels WHERE enabled AND deleted_at IS NULL")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, kind string
		var sealed []byte
		if err = rows.Scan(&id, &kind, &sealed); err != nil {
			return err
		}
		if _, err = s.Cipher.Open(id, kind, sealed); err != nil {
			return ErrUnavailable
		}
	}
	return rows.Err()
}

type Delivery struct {
	ID          string     `json:"id"`
	EventID     string     `json:"alert_event_id"`
	ChannelID   string     `json:"channel_id"`
	Type        string     `json:"channel_type"`
	Status      string     `json:"status"`
	Attempts    int        `json:"attempt_count"`
	LastError   *string    `json:"last_error"`
	NextAttempt *time.Time `json:"next_attempt_at"`
	LastAttempt *time.Time `json:"last_attempt_at"`
	AcceptedAt  *time.Time `json:"accepted_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (s *Store) Deliveries(ctx context.Context, eventID string, limit, offset int) ([]Delivery, error) {
	rows, err := s.Pool.Query(ctx, `SELECT id::text,alert_event_id::text,channel_id::text,channel_type,status,attempt_count,last_error,next_attempt_at,last_attempt_at,accepted_at,created_at FROM alert_deliveries WHERE ($1='' OR alert_event_id=NULLIF($1,'')::uuid) ORDER BY created_at DESC,id DESC LIMIT $2 OFFSET $3`, eventID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Delivery{}
	for rows.Next() {
		var d Delivery
		if err = rows.Scan(&d.ID, &d.EventID, &d.ChannelID, &d.Type, &d.Status, &d.Attempts, &d.LastError, &d.NextAttempt, &d.LastAttempt, &d.AcceptedAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// DeliverOne bloqueia canal e entrega para serializar envio, pausa e arquivo.
func (s *Store) DeliverOne(ctx context.Context, send alerts.Send) (bool, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var id, channel, kind string
	var sealed, body []byte
	var attempt int
	err = tx.QueryRow(ctx, `SELECT d.id::text,d.channel_id::text,d.channel_type,d.config_encrypted,d.payload,d.attempt_count FROM alert_deliveries d JOIN alert_channels c ON c.id=d.channel_id WHERE d.status='pending' AND d.next_attempt_at<=now() AND c.enabled AND c.deleted_at IS NULL ORDER BY d.next_attempt_at,d.id LIMIT 1 FOR UPDATE OF c,d SKIP LOCKED`).Scan(&id, &channel, &kind, &sealed, &body, &attempt)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	config, err := s.Cipher.Open(channel, kind, sealed)
	outcome := alerts.Outcome{Code: "config_error"}
	if err == nil {
		outcome = send(ctx, kind, config, body, id)
	}
	if err = ctx.Err(); err != nil {
		return false, err
	}
	attempt++
	status := "failed"
	var next *time.Time
	var accepted *time.Time
	var code *string
	now := time.Now().UTC()
	if outcome.Accepted {
		status = "sent"
		accepted = &now
	} else {
		v := alerts.SafeCode(outcome.Code)
		code = &v
		if outcome.Retry && attempt < 5 {
			status = "pending"
			v := now.Add(alerts.RetryDelay(attempt))
			next = &v
		}
	}
	_, err = tx.Exec(ctx, "UPDATE alert_deliveries SET status=$2,attempt_count=$3,last_error=$4,next_attempt_at=$5,last_attempt_at=clock_timestamp(),accepted_at=$6,updated_at=clock_timestamp() WHERE id=$1", id, status, attempt, code, next, accepted)
	if err != nil {
		return false, err
	}
	if err = recordDeliveryMetrics(ctx, tx, outcome.Accepted); err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}

func enqueueAlert(ctx context.Context, tx pgx.Tx, eventID string, endpoint Endpoint, check Check, rule alerts.Rule) error {
	p := alerts.Payload{EventID: eventID, Event: rule.Type, Severity: rule.Severity, Timestamp: check.Result.CheckedAt, Certificate: check.Result.Certificate, DaysRemaining: check.Result.DaysRemaining, Code: check.Result.Code}
	p.Endpoint.Name = endpoint.Name
	p.Endpoint.Hostname = endpoint.Hostname
	p.Endpoint.Port = endpoint.Port
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	// FOR SHARE fecha a corrida com pausa/arquivo enquanto o evento é enfileirado.
	_, err = tx.Exec(ctx, `INSERT INTO alert_deliveries(alert_event_id,channel_id,channel_type,config_encrypted,payload) SELECT $1,id,type,config_encrypted,$2 FROM alert_channels WHERE enabled AND deleted_at IS NULL FOR SHARE`, eventID, body)
	return err
}
