package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/franklinlem/certguard/api/internal/alerts"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/jackc/pgx/v5"
	"time"
)

type Alert struct {
	ID           string             `json:"id"`
	EndpointID   string             `json:"endpoint_id"`
	CheckID      *string            `json:"check_id"`
	EndpointName string             `json:"endpoint_name"`
	Hostname     string             `json:"hostname"`
	Port         int                `json:"port"`
	Type         string             `json:"event_type"`
	Severity     string             `json:"severity"`
	Threshold    int                `json:"threshold_days"`
	Snapshot     certificate.Result `json:"snapshot"`
	CreatedAt    time.Time          `json:"created_at"`
	ResolvedAt   *time.Time         `json:"resolved_at"`
	Resolution   *string            `json:"resolution"`
}
type AlertSummary struct {
	Active   int `json:"active"`
	Critical int `json:"critical"`
	Resolved int `json:"resolved"`
	Info     int `json:"info"`
}
type AlertPage struct {
	Items       []Alert      `json:"items"`
	Summary     AlertSummary `json:"summary"`
	Limit       int          `json:"limit"`
	Offset      int          `json:"offset"`
	GeneratedAt time.Time    `json:"generated_at"`
}

// recordAlert roda sob o lock do endpoint e compartilha o commit do check.
func recordAlert(ctx context.Context, tx pgx.Tx, endpoint Endpoint, check Check) error {
	var activeID, key, host string
	err := tx.QueryRow(ctx, "SELECT id::text,dedupe_key,hostname FROM alert_events WHERE endpoint_id=$1 AND resolved_at IS NULL", endpoint.ID).Scan(&activeID, &key, &host)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	rule := alerts.Evaluate(check.Result, endpoint.Warning, endpoint.Critical)
	if rule != nil && activeID != "" && key == alerts.Key(check.Result, *rule) {
		return nil
	}
	if activeID != "" {
		reason := "changed"
		if rule == nil {
			reason = "recovered"
		}
		if host != endpoint.Hostname {
			reason = "target_changed"
		}
		if _, err = tx.Exec(ctx, "UPDATE alert_events SET resolved_at=clock_timestamp(),resolution=$2 WHERE id=$1", activeID, reason); err != nil {
			return err
		}
	}
	if rule == nil {
		if activeID == "" || host != endpoint.Hostname {
			return nil
		}
		rule = &alerts.Rule{Type: "endpoint_recovered", Severity: "info"}
	}
	raw, err := json.Marshal(check.Result)
	if err != nil {
		return err
	}
	var eventID string
	err = tx.QueryRow(ctx, `INSERT INTO alert_events(endpoint_id,check_id,endpoint_name,hostname,port,event_type,severity,threshold_days,dedupe_key,snapshot,resolved_at,resolution)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,CASE WHEN $7='info' THEN clock_timestamp() ELSE NULL END,CASE WHEN $7='info' THEN 'recovered' ELSE NULL END) RETURNING id::text`, endpoint.ID, check.ID, endpoint.Name, endpoint.Hostname, endpoint.Port, rule.Type, rule.Severity, rule.Threshold, alerts.Key(check.Result, *rule), raw).Scan(&eventID)
	if err != nil {
		return err
	}
	return enqueueAlert(ctx, tx, eventID, endpoint, check, *rule)
}

func (s *Store) Alerts(ctx context.Context, state, endpointID string, limit, offset int) (AlertPage, error) {
	out := AlertPage{Items: []Alert{}, Limit: limit, Offset: offset}
	if (state != "all" && state != "active" && state != "resolved") || limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		return out, ErrInvalid
	}
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return out, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	// Indicadores sobre todos os eventos, ou sobre o endpoint solicitado, sem filtro de estado/página.
	filter := ` WHERE ($1='' OR endpoint_id=NULLIF($1,'')::uuid) `
	if err = tx.QueryRow(ctx, `SELECT now(),count(*) FILTER(WHERE resolved_at IS NULL),count(*) FILTER(WHERE resolved_at IS NULL AND severity='critical'),count(*) FILTER(WHERE resolved_at IS NOT NULL AND severity<>'info'),count(*) FILTER(WHERE severity='info') FROM alert_events`+filter, endpointID).Scan(&out.GeneratedAt, &out.Summary.Active, &out.Summary.Critical, &out.Summary.Resolved, &out.Summary.Info); err != nil {
		return out, err
	}
	rows, err := tx.Query(ctx, `SELECT to_jsonb(a) FROM alert_events a`+filter+` AND ($2='all' OR ($2='active' AND resolved_at IS NULL) OR ($2='resolved' AND resolved_at IS NOT NULL)) ORDER BY created_at DESC,id DESC LIMIT $3 OFFSET $4`, endpointID, state, limit, offset)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var a Alert
		if err = rows.Scan(&raw); err != nil {
			return out, err
		}
		if err = json.Unmarshal(raw, &a); err != nil {
			return out, err
		}
		out.Items = append(out.Items, a)
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
