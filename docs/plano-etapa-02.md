# Plano — Etapa 02: acesso autenticado e verificações persistidas

Data: 12/10/2026. Continuidade da fundação, sem implantação ou alteração da VPS.

## Escopo
1. Migração 002 para administrador, sessões persistidas e arquivamento de endpoints, preservando histórico e migração 001.
2. Bootstrap administrativo por CLI e senha via stdin, bcrypt com custo 12; sem cadastro público ou credenciais padrão.
3. Login, logout e sessão atual: cookie HttpOnly/SameSite=Strict/Secure fora do ambiente local, validade absoluta de 8h, token aleatório com apenas hash no banco. CSRF por Origin obrigatório e token específico da sessão para mutações.
4. Limites de tentativa de login com memória limitada, corpo JSON limitado, mensagens sem detalhes sensíveis; sem confiar em X-Forwarded-For arbitrário.
5. CRUD autenticado de endpoints, hostname normalizado, porta 443, validação de limiares/intervalo, arquivamento sem perda de histórico.
6. Check manual com motor existente, concorrência global limitada, lock de endpoint no PostgreSQL, cooldown e persistência atômica do resultado e timestamp. Bloqueio SSRF mantido em todo check.
7. Consulta paginada de endpoints e histórico; interface mínima real para login, cadastro, edição, arquivamento, execução manual e resultados.
8. Testes de sessão/CSRF, autorização, limites, validação de domínio e integração PostgreSQL incluindo persistência/locks. CI, documentação e PR de revisão.

## Decisões e limites
Uma instalação possui um administrador inicial e um conjunto compartilhado de endpoints. Sem multi-tenant, signup público, reset por e-mail, scheduler ou alertas neste incremento. Admin criado por comando explícito; não gerar senha real nem publicá-la no GitHub. Check síncrono com deadline total, sem fila persistida; ampliar junto ao scheduler.

## Critérios de aceite
- Rotas de negócio negadas sem sessão; sessão revogada no logout.
- Login e mutações rejeitam origem inesperada; mutações autenticadas exigem token CSRF.
- Banco não armazena senha ou token de sessão em claro.
- Endpoints privados/URLs/portas proibidas rejeitados, e DNS revalidado pelo motor ao verificar.
- Check e timestamp salvos atomicamente; checks duplicados concorrentes bloqueados.
- Interface comunica falhas e estado real em pt-BR, sem monitoramento periódico fictício.
- Testes, lint, build e auditorias aprovados. PR publicado, sem merge automático ou deploy.
