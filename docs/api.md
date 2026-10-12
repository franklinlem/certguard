# API nesta etapa

`GET /health`: 200 se processo responde. `GET /ready`: 200 se PostgreSQL responde ao Ping, 503 caso contrário. Não revela URL, credenciais ou erro interno.

O contrato implementado está em openapi.yaml. Rotas `/api/v1` do planejamento ainda não foram implementadas. Nenhum endpoint de verificação arbitrária será exposto sem autenticação e limites de uso.

CLI: `certguard check hostname`, porta fixa 443. JSON inclui hostname, checked_at, duration_ms, status, code e, quando disponíveis, resolved_ip, tls_version, cipher_suite, certificate, chain, days_remaining, hostname_valid e chain_valid.

Códigos: success, blocked_target, dns_error, connection_error, connection_timeout, connection_refused, tls_error, certificate_expired, certificate_not_yet_valid, hostname_mismatch, untrusted_chain, internal_error. `connection_error` complementa o planejamento para falhas de rede que não são recusa nem timeout.
