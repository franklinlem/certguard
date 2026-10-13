package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

// CheckDue reivindica apenas um registro, sem bloquear endpoints já em execução.
// O lock também protege contra checks manuais, alterações e arquivamento.
func (s *Store) CheckDue(ctx context.Context, inspect Inspect) (bool, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	endpoint, err := scan(tx.QueryRow(ctx, "SELECT "+columns+` FROM monitored_endpoints
 WHERE enabled AND deleted_at IS NULL AND next_check_at<=now()
 AND (last_check_at IS NULL OR last_check_at<=now()-interval '30 seconds')
 ORDER BY next_check_at,id LIMIT 1 FOR UPDATE SKIP LOCKED`))
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_, err = saveCheck(ctx, tx, endpoint, inspect)
	return err == nil, err
}

// PruneHistory exclui um lote de resultados antigos e mantém o mais recente
// de cada endpoint. O desempate por UUID torna a preservação determinística.
func (s *Store) PruneHistory(ctx context.Context, days, batch int) (int64, error) {
	if days < 30 || days > 36500 || batch < 1 || batch > 1000 {
		return 0, ErrInvalid
	}
	tag, err := s.Pool.Exec(ctx, `WITH candidates AS (
 SELECT c.id FROM certificate_checks c
 WHERE c.checked_at < now()-$1*interval '1 day'
 AND EXISTS (SELECT 1 FROM certificate_checks newer WHERE newer.endpoint_id=c.endpoint_id
   AND (newer.checked_at,newer.id)>(c.checked_at,c.id))
 ORDER BY c.checked_at,c.id LIMIT $2 FOR UPDATE OF c SKIP LOCKED
 ) DELETE FROM certificate_checks c USING candidates d WHERE c.id=d.id`, days, batch)
	return tag.RowsAffected(), err
}
