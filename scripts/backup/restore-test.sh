#!/usr/bin/env bash
# Restaura somente em PostgreSQL descartável, sem redes/volumes existentes.
set -euo pipefail
umask 077
fail() { printf 'Ensaio CertGuard: %s\n' "$1" >&2; exit 1; }
[[ $(hostname) == frank-note2 ]] || fail 'servidor de ensaio incorreto'
CG_FILE=${1:?Informe o arquivo .dump.gpg}
CG_MOUNT=${CERTGUARD_BACKUP_MOUNT:-/srv/storage}
[[ "$CG_MOUNT" == /* ]] || fail 'caminho do HD deve ser absoluto'
CG_ROOT="$CG_MOUNT/backups/certguard"
CG_FILE=$(realpath -e -- "$CG_FILE")
case "$CG_FILE" in "$CG_ROOT/"*/certguard-*.dump.gpg) ;; *) fail 'arquivo fora da pasta CertGuard' ;; esac
mountpoint -q "$CG_MOUNT" || fail 'HD não montado'
[[ $(findmnt -n -o UUID -T "$CG_MOUNT") == "$(cat "$HOME/.local/share/certguard-backup/storage.uuid")" ]] || fail 'HD incorreto'
[[ -f "$CG_FILE.sha256" ]] || fail 'checksum ausente'
(cd "$(dirname "$CG_FILE")"; sha256sum --check --status "$(basename "$CG_FILE").sha256") || fail 'checksum inválido'
CG_CONTAINER="certguard-restore-$(date -u +%Y%m%d%H%M%S)-$$"
CG_CREATED=false
cleanup() {
  CG_STATUS=$?
  if [[ "$CG_CREATED" == true ]]; then docker rm -f "$CG_CONTAINER" >/dev/null; fi
  exit "$CG_STATUS"
}
trap cleanup EXIT
# O único armazenamento é tmpfs; não usa volumes persistentes nem portas.
docker run -d --name "$CG_CONTAINER" --network none --memory 512m --cpus 1 \
  --pids-limit 128 --security-opt no-new-privileges=true \
  --tmpfs /var/lib/postgresql/data:rw,size=512m \
  --tmpfs /var/run/postgresql:rw,size=16m \
  -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_USER=certguard \
  -e POSTGRES_DB=certguard_restore postgres:17-alpine >/dev/null
CG_CREATED=true
CG_READY=false
for ((CG_TRY=0;CG_TRY<60;CG_TRY++)); do
  if docker exec "$CG_CONTAINER" pg_isready -U certguard -d certguard_restore >/dev/null 2>&1; then CG_READY=true; break; fi
  sleep 0.5
done
[[ "$CG_READY" == true ]] || fail 'PostgreSQL descartável não ficou pronto'
# Nunca pede senha em sessão automatizada: o responsável desbloqueia antes.
gpg --homedir "$HOME/.local/share/certguard-backup/gnupg" --batch --pinentry-mode error \
  --decrypt "$CG_FILE" \
  | docker exec -i "$CG_CONTAINER" pg_restore --exit-on-error --no-owner \
      -U certguard -d certguard_restore
CG_RESULT=$(docker exec "$CG_CONTAINER" psql -X -v ON_ERROR_STOP=1 -U certguard -d certguard_restore -Atc \
  "SELECT json_build_object('migrations',(SELECT count(*) FROM schema_migrations),'schema',(SELECT max(version) FROM schema_migrations),'users',(SELECT count(*) FROM users),'endpoints',(SELECT count(*) FROM monitored_endpoints),'checks',(SELECT count(*) FROM certificate_checks),'events',(SELECT count(*) FROM alert_events),'channels',(SELECT count(*) FROM alert_channels),'deliveries',(SELECT count(*) FROM alert_deliveries));")
python3 -c 'import json,sys; d=json.loads(sys.argv[1]); assert d["migrations"]==6 and d["schema"]==6; assert d["users"]>=1; print("Restauração descartável aprovada:",json.dumps(d))' "$CG_RESULT"
