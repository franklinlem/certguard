#!/usr/bin/env bash
# Executado no frank-note2; não altera a VPS nem outros backups.
set -euo pipefail
umask 077
fail() { printf 'Backup CertGuard: %s\n' "$1" >&2; exit 1; }
[[ $(hostname) == frank-note2 ]] || fail 'servidor de destino incorreto'
CG_STATE="$HOME/.local/share/certguard-backup"
CG_MOUNT=${CERTGUARD_BACKUP_MOUNT:-/srv/storage}
[[ "$CG_MOUNT" == /* ]] || fail 'caminho do HD deve ser absoluto'
CG_ROOT="$CG_MOUNT/backups/certguard"
CG_ENV=${1:-homologacao}
case "$CG_ENV" in
  homologacao) CG_PROJECT=rk48swz6mdqduoyyhfnqhylp ;;
  producao)
    [[ -f "$CG_STATE/production-resource" ]] || fail 'recurso de produção ainda não cadastrado'
    CG_PROJECT=$(cat "$CG_STATE/production-resource")
    [[ "$CG_PROJECT" != rk48swz6mdqduoyyhfnqhylp ]] || fail 'produção não pode usar homologação'
    ;;
  *) fail 'ambiente inválido' ;;
esac
[[ "$CG_PROJECT" =~ ^[a-z0-9][a-z0-9-]{0,63}$ ]] || fail 'identidade de recurso inválida'
mountpoint -q "$CG_MOUNT" || fail 'HD de backup não está montado'
[[ -s "$CG_STATE/storage.uuid" ]] || fail 'identidade do HD ausente'
[[ $(findmnt -n -o UUID -T "$CG_MOUNT") == "$(cat "$CG_STATE/storage.uuid")" ]] || fail 'HD de backup diferente do configurado'
[[ -d "$CG_ROOT" && -O "$CG_ROOT" && ! -L "$CG_ROOT" ]] || fail 'diretório de backup inválido'
CG_FREE=$(df -B1 --output=avail "$CG_MOUNT" | tail -1 | tr -d ' ')
[[ "$CG_FREE" =~ ^[0-9]+$ && "$CG_FREE" -gt 268435456 ]] || fail 'espaço livre insuficiente'
[[ -s "$CG_STATE/recipient" ]] || fail 'chave pública de backup não configurada'
CG_RECIPIENT=$(cat "$CG_STATE/recipient")
[[ "$CG_RECIPIENT" =~ ^[A-F0-9]{40,64}$ ]] || fail 'fingerprint de chave inválido'
install -d -m 700 "$HOME/.local/state/certguard-backup"
exec 9> "$HOME/.local/state/certguard-backup/backup.lock"
flock -n 9 || fail 'outro backup CertGuard está em execução'
CG_DIR="$CG_ROOT/$CG_ENV"
[[ ! -L "$CG_DIR" ]] || fail 'diretório de ambiente não pode ser link'
install -d -m 700 "$CG_DIR"
CG_TMP=$(mktemp "$CG_DIR/.certguard-$(date -u +%Y%m%dT%H%M%SZ)-XXXXXX")
cleanup() {
  CG_STATUS=$?
  [[ -z "$CG_TMP" ]] || rm -f -- "$CG_TMP"
  exit "$CG_STATUS"
}
trap cleanup EXIT
ssh -n -i "$HOME/.ssh/contabo_backup" -o IdentitiesOnly=yes -o BatchMode=yes \
  -o StrictHostKeyChecking=yes -o ConnectTimeout=10 \
  -o ServerAliveInterval=15 -o ServerAliveCountMax=3 \
  franklin@158.220.120.2 \
  "docker exec ${CG_PROJECT}-db-1 pg_dump --format=custom --no-owner --username=certguard --dbname=certguard" \
  | gpg --homedir "$CG_STATE/gnupg" --batch --yes --no-encrypt-to \
      --trust-model always --recipient "$CG_RECIPIENT" --encrypt --output "$CG_TMP"
[[ -s "$CG_TMP" ]] || fail 'arquivo de backup vazio'
CG_NAME="$(basename "$CG_TMP")"
CG_NAME="${CG_NAME#.}.dump.gpg"
CG_FINAL="$CG_DIR/$CG_NAME"
# Hard link publica atomicamente e falha se o destino já existir.
ln -- "$CG_TMP" "$CG_FINAL"
rm -- "$CG_TMP"
CG_TMP=''
(cd "$CG_DIR"; sha256sum "$CG_NAME" > "$CG_NAME.sha256")
printf 'Backup CertGuard concluído (%s): %s\n' "$CG_ENV" "$CG_FINAL"
