# API implementada — etapa 04

Todas as respostas de erro usam `{ "error": "mensagem em português" }`, sem SQL, secrets ou dados internos. Corpos JSON limitados a 8 KiB, campos desconhecidos rejeitados.

| Método | Rota | Acesso / comportamento |
|---|---|---|
| GET | /health | Público: processo disponível |
| GET | /ready | Público: PostgreSQL responde |
| POST | /api/v1/auth/login | Origin exata; e-mail/senha; retorna usuário, csrf_token e cookie |
| GET | /api/v1/auth/me | Sessão válida; retorna usuário e csrf_token |
| POST | /api/v1/auth/logout | Sessão, Origin e CSRF; revoga sessão |
| GET | /api/v1/dashboard | Indicadores globais e endpoints paginados; sessão obrigatória |
| GET/POST | /api/v1/endpoints | Lista paginada / cadastra |
| GET/PATCH/DELETE | /api/v1/endpoints/{id} | Consulta / alteração parcial / arquivamento |
| POST | /api/v1/endpoints/{id}/check | Inspeção TLS síncrona, resultado persistido |
| GET | /api/v1/endpoints/{id}/checks | Histórico, inclusive de endpoint arquivado |

Rotas de endpoints exigem sessão; mutações exigem Origin e X-CSRF-Token. Sem bearer tokens nesta etapa. Paginação: limit=1..100 (padrão 50), offset=0..100000 (padrão 0). Listas retornam items, limit e offset.

Cadastro: name, hostname; opcionais port=443, enabled=true, interval_minutes=360, warning_days=30 e critical_days=7. Intervalo 5..10080, warning > critical >=0, warning <=36500. O intervalo controla a cadência automática quando o agendador estiver habilitado, acrescido de jitter de até 10%. DELETE arquiva e permite recadastrar o mesmo hostname como novo endpoint.

Check retorna 201 com id, endpoint_id e result. Resultado de certificado inválido também retorna 201: o check foi registrado; consulte result.code/status. Conflito/lock/cooldown/pausado: 409; limite global de concorrência: 429. Credenciais/sessão inválidas: 401; origem/CSRF: 403; banco indisponível: 503.

Detalhes e schemas: openapi.yaml. Códigos do motor: success, blocked_target, dns_error, connection_error, connection_timeout, connection_refused, tls_error, certificate_expired, certificate_not_yet_valid, hostname_mismatch, untrusted_chain e internal_error.

Login e `/auth/me` acrescentam `background`: `scheduler_enabled`, `check_workers`, `retention_enabled`, `retention_days`. São configurações desta instância, não uma medição de saúde do executor. Endpoints incluem `next_check_at` anulável. Check manual também atualiza a agenda; pausa/arquivo removem a próxima execução. Retenção habilitada pode reduzir o histórico, preservando sempre o último resultado.

Dashboard: `summary` contém total, enabled, paused, healthy, warning, critical, errors, unchecked e expiring_30_days. `items` segue o endpoint acrescido de `latest` (Result ou null); `generated_at` identifica o snapshot. Limite/offset iguais aos endpoints. Indicadores incluem pausados e excluem arquivados. Status do último check do hostname/porta atuais; vencimento entre agora e 30 dias, excluindo expirados. Não dispara checks. Veja ADR-006.
