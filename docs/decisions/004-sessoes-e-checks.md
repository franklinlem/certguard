# ADR-004 — Sessões e verificações manuais persistidas

Status: implementado, em validação. Data: 12/10/2026.

## Acesso
Instalação de um administrador, criado por CLI com senha via stdin. Não existem cadastro público, recuperação por e-mail nem rotação de senha via UI nesta etapa. Senhas de 12–72 bytes, bcrypt custo 12. Login desconhecido usa hash fictício de mesmo custo para reduzir diferenças de tempo. Duas comparações de senha simultâneas por processo.

Sessão de 256 bits aleatórios, validade absoluta de 8h, hash SHA-256 armazenado no PostgreSQL. Cookie __Host-certguard_session quando Secure (certguard_session em HTTP local), host-only, HttpOnly, SameSite=Strict; Secure fora de desenvolvimento local. Limite de dez sessões ativas por usuário, limpeza de expiradas ao criar sessão; logout revoga a sessão atual. Não há logout global ou renovação deslizante.

Origem exata PUBLIC_APP_URL em todas as mutações, inclusive login. Sessões autenticadas exigem X-CSRF-Token, derivado do segredo de sessão com prefixo próprio. Token CSRF entregue por login e /auth/me, mantido apenas em memória no frontend. Sem CORS e sem localStorage para tokens.

Rate limiting em memória: 60 tentativas/min global, dez por endereço em 15min e cinco por e-mail em 15min. Conta tentativas bem e malsucedidas. Capacidade de 4096 chaves e expiração. Não confiar em X-Forwarded-For: atrás do proxy, IP pode agrupar todos os clientes. Reinício limpa contadores; replicação ou exposição pública exige limite compartilhado e política confiável de proxy. Esta etapa pressupõe uma instância.

## Endpoints e histórico
Validação sintática de hostname no cadastro e proteção DNS/IP antes de toda conexão no motor. Nome DNS que resolva para rede privada pode ser armazenado, porém seu check resulta em blocked_target sem conexão. Porta exclusivamente 443. PATCH aplica somente campos enviados sob lock; arquivamento preserva checks. Recadastro do hostname arquivado recebe novo id. Histórico original continua disponível por id na API, mas a UI não lista arquivados nesta etapa.

Cada check manual mantém transação e lock FOR UPDATE NOWAIT do endpoint durante até dez segundos de inspeção. Resultado e last_check_at são gravados atomicamente. Concorrência máxima quatro por processo, cooldown de 30s por endpoint entre checks concluídos. Pausa impede verificações; concorrência entre processos é serializada pelo lock no banco. Sem fila, retry automático ou agendamento. CHECK_TIMEOUT do motor continua dez segundos; HTTP tem deadline de 15s e write timeout de 20s.

Migração 002 preserva a 001 e aplica alterações transacionalmente sob o mesmo advisory lock. O runner versionado continua mínimo; introdução de goose foi avaliada e adiada para não criar dois históricos de versão neste incremento. Não executar rollback destrutivo automático.

Referências: https://pkg.go.dev/golang.org/x/crypto/bcrypt e https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html.
