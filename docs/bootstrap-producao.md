# Primeiro administrador de produção

Executar somente após migração, API e HTTPS validados em `https://certguard.franklem.uk`. A senha é exclusiva de produção e não é a senha PostgreSQL. Não enviar senha na conversa nem versionar arquivos com credenciais.

No terminal local, iniciar `ssh contabo` e executar o bloco abaixo. Ele confirma o servidor e seleciona somente a API do recurso de produção. A CLI recebe a senha pela entrada padrão, sem registrá-la em argumentos ou histórico.

```bash
(
  set +x
  set -eu
  test "$(hostname)" = vmi3570746
  CG_PROD_API=$(docker ps --filter label=com.docker.compose.project=oiiuvtwkobgvmjeo9j1odxcp --filter label=com.docker.compose.service=api --format '{{.ID}}')
  test -n "$CG_PROD_API"
  test "$(printf '%s\n' "$CG_PROD_API" | wc -l)" -eq 1
  IFS= read -r -s -p 'Senha do administrador de produção (mínimo 10 caracteres): ' CG_PROD_PASSWORD
  printf '\n'
  IFS= read -r -s -p 'Repita a senha: ' CG_PROD_CONFIRM
  printf '\n'
  test "$CG_PROD_PASSWORD" = "$CG_PROD_CONFIRM" || { echo 'Senhas diferentes.'; exit 1; }
  printf '%s' "$CG_PROD_PASSWORD" | docker exec -i "$CG_PROD_API" certguard admin-create franklinlem@gmail.com
  unset CG_PROD_PASSWORD CG_PROD_CONFIRM
)
```

O comando recusa um segundo administrador inicial. Não apagar usuários ou banco para contornar essa proteção. Após criar, validar login e logout em produção. Não reutilizar a senha de homologação.
