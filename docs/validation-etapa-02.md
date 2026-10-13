# Validação — etapa 02

Data: 12/10/2026. Código validado: `a18fbfdea0a1fa2689c97e4089de2a904510ea6c`.
PR: https://github.com/franklinlem/certguard/pull/6.

## Evidências
- [Validação de backend, frontend e Docker/E2E](https://github.com/franklinlem/certguard/actions/runs/38294848666).
- [Auditorias de código, dependências e imagens](https://github.com/franklinlem/certguard/actions/runs/38294848853).
- Testes Go com race detector, go vet, Staticcheck, cobertura mínima dos pacotes core, fuzzing e build.
- PostgreSQL 17: migrações idempotentes, bootstrap único, credenciais, expiração/hash de sessão, duplicidade, locks concorrentes, cooldown, pausa, rollback e histórico preservado após arquivamento.
- Fluxo HTTP com banco real: login, cookie, CSRF, Origin, autorização, cadastro, PATCH estrito, check persistido, histórico, arquivo, logout e rate limiting. Motor substituído somente no teste HTTP para não abrir conexões externas.
- Frontend: lint, TypeScript, testes de sessão e disparo de check com CSRF, build e npm audit.
- Docker Compose: build, inicialização, migrations, saúde/readiness e proxy API.
- Playwright: um cenário E2E aprovado, percorrendo login, cadastro, edição, pausa, histórico vazio, arquivamento e logout em navegador Chromium. Administrador e banco usados são descartáveis.
- OpenAPI: referências internas resolvidas e schemas sem required vazio. Não foi executado um validador externo completo do contrato.

## Segurança e limites
Govulncheck não identificou vulnerabilidades alcançáveis pelo código. O módulo x/crypto contém o pacote OpenPGP descontinuado, mas ele não é importado nem utilizado pelo CertGuard. As verificações Trivy de código/imagens passaram nos gates HIGH/CRITICAL configurados; nas imagens, vulnerabilidades sem correção são excluídas do gate. Isso não equivale a ausência absoluta de vulnerabilidades.

Docker/PostgreSQL não estão instalados localmente; integração e E2E foram executados nos runners efêmeros do GitHub. Testes locais deixam explícito quando não há TEST_DATABASE_URL.

Um commit posterior somente de documentação pode atualizar este relatório e corrigir o contrato sem mudar o código validado acima. A primeira execução da CI foi cancelada quando a revisão de PATCH foi publicada; as evidências apontam para a revisão final do código.

Sem scheduler, alertas, recuperação de senha ou listagem de arquivados na UI. Sem merge automático. Nenhuma alteração ou implantação na VPS, homologação ou produção.
