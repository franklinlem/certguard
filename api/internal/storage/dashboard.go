package storage

import (
	"context"
	"encoding/json"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/jackc/pgx/v5"
	"time"
)

type EndpointView struct {
	Endpoint
	Latest *certificate.Result `json:"latest"`
}
type Summary struct {
	Total      int `json:"total"`
	Enabled    int `json:"enabled"`
	Paused     int `json:"paused"`
	Healthy    int `json:"healthy"`
	Warning    int `json:"warning"`
	Critical   int `json:"critical"`
	Errors     int `json:"errors"`
	Unchecked  int `json:"unchecked"`
	Expiring30 int `json:"expiring_30_days"`
}
type Dashboard struct {
	Summary     Summary        `json:"summary"`
	Items       []EndpointView `json:"items"`
	GeneratedAt time.Time      `json:"generated_at"`
	Limit       int            `json:"limit"`
	Offset      int            `json:"offset"`
}

const latestJoin = ` LEFT JOIN LATERAL (
 SELECT c.metadata,c.status FROM certificate_checks c
 WHERE c.endpoint_id=e.id AND c.metadata->>'hostname'=e.hostname
 AND c.metadata->>'port'=e.port::text
 ORDER BY c.checked_at DESC,c.id DESC LIMIT 1
 ) latest ON true `

// Dashboard lê totais globais e página no mesmo snapshot, sem executar checks.
func (s *Store) Dashboard(ctx context.Context, limit, offset int) (Dashboard, error) {
	out := Dashboard{Items: []EndpointView{}, Limit: limit, Offset: offset}
	if limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		return out, ErrInvalid
	}
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return out, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	err = tx.QueryRow(ctx, `SELECT now(),count(*),count(*) FILTER(WHERE e.enabled),count(*) FILTER(WHERE NOT e.enabled),
 count(*) FILTER(WHERE latest.status='healthy'),count(*) FILTER(WHERE latest.status='warning'),
 count(*) FILTER(WHERE latest.status='critical'),count(*) FILTER(WHERE latest.status='error'),
 count(*) FILTER(WHERE latest.status IS NULL),
 count(*) FILTER(WHERE (latest.metadata->'certificate'->>'not_after')::timestamptz>=now()
 AND (latest.metadata->'certificate'->>'not_after')::timestamptz<=now()+interval '30 days')
 FROM monitored_endpoints e`+latestJoin+`WHERE e.deleted_at IS NULL`).Scan(&out.GeneratedAt, &out.Summary.Total, &out.Summary.Enabled, &out.Summary.Paused, &out.Summary.Healthy, &out.Summary.Warning, &out.Summary.Critical, &out.Summary.Errors, &out.Summary.Unchecked, &out.Summary.Expiring30)
	if err != nil {
		return out, err
	}
	rows, err := tx.Query(ctx, `SELECT to_jsonb(e),latest.metadata FROM
 (SELECT * FROM monitored_endpoints WHERE deleted_at IS NULL ORDER BY created_at DESC,id LIMIT $1 OFFSET $2) e`+latestJoin+`ORDER BY e.created_at DESC,e.id`, limit, offset)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw, result []byte
		var v EndpointView
		if err = rows.Scan(&raw, &result); err != nil {
			return out, err
		}
		if err = json.Unmarshal(raw, &v.Endpoint); err != nil {
			return out, err
		}
		if result != nil {
			if err = json.Unmarshal(result, &v.Latest); err != nil {
				return out, err
			}
		}
		out.Items = append(out.Items, v)
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
