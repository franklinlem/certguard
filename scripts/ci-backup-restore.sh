#!/usr/bin/env bash
# Exclusivo do projeto efêmero da CI; nunca restaura em banco existente.
set -euo pipefail
if [[ "${CI:-}" != true || "${GITHUB_ACTIONS:-}" != true ]]; then
  echo 'Este ensaio só pode rodar no ambiente descartável do GitHub Actions.' >&2
  exit 2
fi
umask 077
CERTGUARD_BACKUP_DIR=$(mktemp -d)
trap 'rm -rf -- "$CERTGUARD_BACKUP_DIR"' EXIT
compose=(docker compose -p certguard-ci -f deploy/compose.yaml)
"${compose[@]}" exec -T db pg_dump -U certguard -d certguard -Fc > "$CERTGUARD_BACKUP_DIR/database.dump"
# createdb falha se o destino existir; não há DROP, --clean ou sobrescrita.
"${compose[@]}" exec -T db createdb -U certguard certguard_restore_ci
"${compose[@]}" exec -T db pg_restore -U certguard --no-owner --exit-on-error -d certguard_restore_ci < "$CERTGUARD_BACKUP_DIR/database.dump"
query="SELECT json_build_array((SELECT count(*) FROM schema_migrations),(SELECT max(version) FROM schema_migrations),(SELECT count(*) FROM users),(SELECT count(*) FROM sessions),(SELECT count(*) FROM monitored_endpoints),(SELECT count(*) FROM certificate_checks),(SELECT count(*) FROM alert_events),(SELECT count(*) FROM alert_channels),(SELECT count(*) FROM alert_deliveries),(SELECT checks_total FROM operational_metrics),(SELECT duration_buckets FROM operational_metrics));"
"${compose[@]}" exec -T db psql -X -v ON_ERROR_STOP=1 -U certguard -d certguard -Atc "$query" > "$CERTGUARD_BACKUP_DIR/source.txt"
"${compose[@]}" exec -T db psql -X -v ON_ERROR_STOP=1 -U certguard -d certguard_restore_ci -Atc "$query" > "$CERTGUARD_BACKUP_DIR/restored.txt"
cmp "$CERTGUARD_BACKUP_DIR/source.txt" "$CERTGUARD_BACKUP_DIR/restored.txt"
echo 'Backup/restore descartável aprovado: schema, contagens e contadores preservados.'
