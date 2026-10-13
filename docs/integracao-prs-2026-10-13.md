# Integração dos PRs — 13/10/2026

Integração autorizada pelo responsável. Todos os PRs 1–11 foram integrados por merge commits, preservando a ancestralidade dos incrementos. Não houve implantação automática ou manual.

## Resultado

Código integrado: `e7970ba18bbcc36ec42e4d8aafcb69d4ad4ed952`, branch `main`, migrações esperadas **1–6**. Nenhum PR aberto após a integração. Este SHA identifica o código validado; o registro documental posterior não altera seu funcionamento.

| PR | Conteúdo | Commit de integração |
|---|---|---|
| [6](https://github.com/franklinlem/certguard/pull/6) | Autenticação, endpoints, checks | a3b6946f7390b75f19a7d62fe47a505f228c8ef4 |
| [7](https://github.com/franklinlem/certguard/pull/7) | Scheduler e retenção | 1e152178596c0d8a267e6760311f8118791446f9 |
| [8](https://github.com/franklinlem/certguard/pull/8) | Dashboard | 4926086f0d32e5642a95a2321b3cad2e95c28d6d |
| [9](https://github.com/franklinlem/certguard/pull/9) | Eventos de alerta | e2b1713c7b23c9eee3abd1bdc8de0bf7e39263ae |
| [10](https://github.com/franklinlem/certguard/pull/10) | Entregas SMTP/webhook | 6077f7562668c7cecc5767832d3c2ea01d35cb45 |
| [11](https://github.com/franklinlem/certguard/pull/11) | Métricas/readiness/hardening | 442ec7d2dc9cb5fdd729a81dd13babda9d36ebf5 |
| [1](https://github.com/franklinlem/certguard/pull/1) | setup-node 7.1.0 | a047e276d19b31e940f709b96442742b7557cbc3 |
| [2](https://github.com/franklinlem/certguard/pull/2) | Alpine 3.24 na API | 4110cc72f9f63a480b3a6736abb605c3a06461be |
| [3](https://github.com/franklinlem/certguard/pull/3) | setup-go 7.0.0 | 7405874971b94809278a83028e0900bae6aaf72f |
| [5](https://github.com/franklinlem/certguard/pull/5) | Node 26.11.1 na imagem de build web | 78bddbbf13b05cf40176c1d922e361b6db3f2bee |
| [4](https://github.com/franklinlem/certguard/pull/4) | checkout 7.0.1 | e7970ba18bbcc36ec42e4d8aafcb69d4ad4ed952 |

## Decisões e compatibilidade

- Bases dos PRs funcionais foram ajustadas para `main` após integração da dependência anterior. Históricos e registros de validação originais foram preservados.
- Os cinco PRs Dependabot receberam o produto completo antes dos testes: a validação antiga somente sobre a fundação não foi considerada suficiente.
- A nova ocorrência de setup-node no job Docker também foi atualizada. O conflito de checkout com setup-go foi resolvido sobre a main atual, preservando as duas atualizações e todas as funcionalidades.
- Ações permanecem fixadas por SHA. Os READMEs oficiais das versões [setup-node](https://github.com/actions/setup-node/tree/v7.1.0), [setup-go](https://github.com/actions/setup-go/tree/v7.0.0) e [checkout](https://github.com/actions/checkout/tree/v7.0.1) foram consultados: os workflows usam runners hospedados Ubuntu, entradas explícitas de versão/cache, sem registry-url, pull_request_target ou workflow_run.
- Node local e instalado nos jobs permanece 22.23.2; a imagem Docker de build usa 26.11.1, testada separadamente pelos builds, auditoria e E2E em Compose. O frontend servido continua estático pelo Nginx; não há Node no runtime web. Não houve atualização automática do Node do servidor.

## Evidências

Código funcional integrado `442ec7d2`: [Validação 38406712397](https://github.com/franklinlem/certguard/actions/runs/38406712397) e [Segurança 38406713578](https://github.com/franklinlem/certguard/actions/runs/38406713578), ambas aprovadas.

Validações dos PRs Dependabot sobre o produto completo, todas aprovadas:

| PR | Validação | Segurança |
|---|---|---|
| 1 | 38406897398 | 38406897726 |
| 2 | 38406906621 | 38406907277 |
| 3 | 38406926151 | 38406925846 |
| 4 (antes do conflito) | 38406916793 | 38406917585 |
| 5 | 38406926670 | 38406926394 |
| 4 (conflito resolvido) | 38407814257 | 38407813817 |

Conjunto final `e7970ba1`: [Validação 38408733689](https://github.com/franklinlem/certguard/actions/runs/38408733689) e [Segurança 38408734284](https://github.com/franklinlem/certguard/actions/runs/38408734284). **Ambas aprovadas no SHA final**, incluindo os jobs backend, frontend e Docker. Esse código é candidato para homologação, ainda sem aceite operacional.

As execuções cobrem Go/race/vet/staticcheck, PostgreSQL/migrações/concorrência, cobertura/fuzzing, leituras de carga sintética, frontend/typecheck/lint/testes/build/audit, imagens, Compose, E2E desktop/mobile e backup/restauração em banco descartável. Execuções intermediárias de validação de main foram canceladas pela concorrência configurada ao chegar o commit seguinte; não são apresentadas como aprovadas. Segurança dos estados intermediários foi executada; o conjunto final é o critério de integração.

## Estado operacional

Consulte [inventário](inventario-homologacao-2026-10-13.md) e [roteiro](homologacao.md). DNS criado pelo responsável; aplicação ainda não implantada. CI não é aceite de homologação. Produção continua condicionada ao aceite documentado do ambiente de testes; qualquer alteração com risco de interrupção/perda de dados exige autorização específica.
