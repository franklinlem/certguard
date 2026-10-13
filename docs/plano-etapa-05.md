# Etapa 05 — motor e histórico de alertas

## Objetivo e base

Implementar a primeira entrega da v0.5.0: avaliação de regras, eventos persistidos, deduplicação, resolução e consulta administrativa. Base: dashboard / PR #8, ainda aberto em 13/10/2026. Nova branch e PR dependente. SMTP, webhook, canais e retries compõem a próxima entrega, após validar as regras; nenhum envio externo nesta etapa.

## Regras e persistência

1. Avaliar somente resultados novos, manuais ou automáticos, dentro da mesma transação do check. Uma falha ao salvar evento reverte resultado/agenda para impedir perda silenciosa de alertas.
2. Atenção/crítico conforme limiares do endpoint, lembrete ao chegar a até três dias dentro da faixa crítica, expirado, certificado futuro, hostname incompatível, cadeia inválida e erros DNS/TCP/TLS/SSRF. Eventos técnicos são críticos; isso não significa que o certificado expirou.
3. Manter no máximo um problema ativo por endpoint. Chave de deduplicação inclui destino, regra/limiar e fingerprint quando disponível. Repetir o mesmo resultado não gera evento adicional. Mudança de regra, limiar ou certificado resolve o problema anterior e abre outro; uma recorrência após resolução abre novo episódio.
4. Resultado saudável e validado resolve o ativo e cria recuperação informativa. Não criar recuperação no primeiro sucesso. Falha de rede não representa recuperação. Os eventos guardam snapshot do resultado e do destino.
5. Pausa mantém o problema ativo identificado como observação histórica. Troca de destino ou arquivamento resolve o anterior com motivo explícito, sem indicar recuperação TLS. Alterar nome não altera a deduplicação; mudança de limiar é avaliada no próximo check.
6. Migração aditiva 004: tabela de eventos, índice único parcial de problema ativo e índices por endpoint/data. Referência ao check usa ON DELETE SET NULL para não bloquear retenção; snapshot persiste. Sem backfill: histórico antigo não vira tempestade de alertas.

## API e interface

- `GET /api/v1/alerts` autenticado, paginado, filtros `state=all|active|resolved` e endpoint UUID opcional. Indicadores globais de problemas ativos/críticos e eventos resolvidos/informativos, sem depender da página.
- Eventos exibem destino observado, regra, severidade, data, estado, motivo de resolução e ligação ao check quando preservado. O painel informa claramente que entrega por e-mail/webhook ainda não está implementada.
- Atualização explícita, estados de carregamento/erro/vazio separados; paginação não esconde ativos por padrão. Conteúdo de certificados renderizado somente como texto. Sem alterações nos controles SSRF, autenticação ou CSRF.

## Validação e aceite

- Unitários das regras: limiares personalizados, três dias, zero/negativo, falhas técnicas, cadeia/hostname e recuperação.
- PostgreSQL: dedupe/reincidência, mudança de certificado/destino, transições, concorrência, rollback atômico, arquivo/pausa, retenção sem perda de evento e filtros/paginação.
- HTTP: sessão obrigatória, UUID/filtros/limites válidos, resposta sem dados internos.
- Frontend e E2E: estados e timeline, sem mensagens externas ou inspeções TLS públicas.
- Go race/vet/staticcheck, frontend tipos/lint/testes/build, PostgreSQL/Docker/Playwright/segurança na CI. Registrar resultados reais.

## Operação e limites

Geração de eventos acompanha checks novos, inclusive quando scheduler está desabilitado e se usa check manual. Não envia notificações, não necessita credenciais novas e não altera a VPS. Retenção de eventos será definida junto do hardening (planejamento global sugere 365 dias); nesta entrega eventos permanecem preservados. Integração dos PRs anteriores e homologação continuam pendentes.
