#!/usr/bin/env bash
# Executa ambientes configurados em sequência, sem interferir em outros cron jobs.
set -uo pipefail
CG_SCRIPT="$HOME/.local/lib/certguard-backup/backup.sh"
CG_STATUS=0
for CG_ENV in homologacao producao; do
  if [[ "$CG_ENV" == producao && ! -s "$HOME/.local/share/certguard-backup/production-resource" ]]; then
    continue
  fi
  if ! bash "$CG_SCRIPT" "$CG_ENV"; then CG_STATUS=1; fi
done
exit "$CG_STATUS"
