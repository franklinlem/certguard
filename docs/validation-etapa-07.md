# Validação da etapa 07

Base: PR #10. Revisão dos PRs registrada em revisao-prs-etapa-07.md, sem merges. Nenhuma alteração na VPS ou mensagem real.

## Local

- Go race, vet e staticcheck aprovados.
- Métricas: auth/fail-closed, no-store, erros genéricos, formato e cardinalidade fixa. SQL aprovado na CI.
- Frontend: tipos/lint/quatorze testes/build aprovados. Script de backup passou na análise sintática; não foi executado localmente. Não há Docker/PostgreSQL locais.

## CI

Código validado: `12678d72451a1b90bdd9333fa928b0e0b00a241a`.

- [Validação #38403767490](https://github.com/franklinlem/certguard/actions/runs/38403767490): Go/PostgreSQL/race/vet/staticcheck, cobertura/fuzzing, smoke concorrente, benchmark, frontend, Docker, dump/restore e E2E desktop/mobile aprovados.
- [Segurança #38403767536](https://github.com/franklinlem/certguard/actions/runs/38403767536): dependências, código e imagens aprovados.
- [PR #11](https://github.com/franklinlem/certguard/pull/11), dependente do PR #10, aberto sem merge.

### Benchmark de leitura

GitHub Actions ubuntu-latest, Go 1.27.2, PostgreSQL 17, 1000 endpoints e 3000 checks sintéticos, 20 operações por benchmark após setup. Valores médios do benchmark Go; sem percentis/SLO, HTTP, login ou inspeção externa.

| Consulta | ns/op | Média aproximada | B/op | Alocações/op |
|---|---:|---:|---:|---:|
| Dashboard (página de 20 + agregados) | 3259075 | 3,26 ms | 28912 | 187 |
| Métricas (snapshot global) | 3038427 | 3,04 ms | 3197 | 45 |

Smoke concorrente: quatro leitores, dez ciclos de dashboard+métricas por leitor, verificando totais/resultados. Não estabelece throughput máximo ou capacidade de escrita.

### Ensaio operacional

Dois E2E (desktop e mobile/dark) aprovados em 4,0s no runner, incluindo ausência de overflow e bloqueio de rotas de métricas pelo proxy. Script CI confirmou dump/restore em banco novo do projeto efêmero: schema, contagens e contadores preservados. Não valida chave de produção ou provedor SMTP real.

O commit posterior apenas registra evidências; código permanece igual ao validado. Nenhum benchmark, backup/restore ou alteração foi executado na VPS.

## Limites

Benchmark de leituras não mede capacidade de escrita, TLS externo, SMTP ou VPS. Contadores globais por banco e sem histórico anterior já removido. Homologação real não executada; roteiro depende de inventário/acesso e autorização para alterações com risco. Retenção de eventos/entregas e rotação assistida de chave ainda pendentes.
