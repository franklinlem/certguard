# Etapa 07 — hardening e preparação da homologação

## Base e revisão dos PRs

Base: etapa 06, PR #10. Em 13/10/2026, PRs #6–#10 abertos, sem conflitos, revisões ou comentários de alteração; implementação validada e último commit de cada branch apenas documental. PRs Dependabot #1–#5 também possuem CI verde sobre main; major upgrades precisam de avaliação integrada, sem presumir que o teste sobre a fundação valida o produto atual. Nenhum merge automático nesta etapa. Registrar relatório com links/SHAs.

## Escopo

1. `/metrics` no formato Prometheus, desabilitado (404) sem METRICS_TOKEN; bearer de pelo menos 32 caracteres, comparação constante, timeout de consulta e respostas sem cache. Sem etiquetas de hostname, e-mail, UUID, URL, segredo ou código remoto para evitar vazamento/cardinalidade ilimitada.
2. Contadores duráveis de checks, falhas, duração/histograma e tentativas/aceites/falhas de entrega, atualizados na transação do trabalho. Migração 006 inicializa a partir dos registros ainda existentes; retenção posterior não reduz contadores. Histogramas cumulativos de duração em segundos (0.1/0.5/1/2.5/5/10/30/+Inf). Dados anteriores já removidos não podem ser reconstruídos.
3. Gauges de endpoints, classificação do último resultado, vencimentos e fila. Coleta somente leitura; não executa checks/envios. Endpoint não publicado pelo proxy web; scrape precisa de acesso interno à API e token, sem expor portas na VPS.
4. `/ready` verifica PostgreSQL e versão completa do schema esperado; rejeita migrações ausentes/incompatíveis sem SQL/segredos na resposta. `/health` continua liveness. Migrações aditivas, sem limpeza ou retrocesso de schema.
5. Índices para último resultado do destino atual e histórico global de entregas. Benchmark PostgreSQL descartável com 1000 endpoints/3000 checks, sem inspeções de rede; smoke concorrente das leituras. Registrar números e limites da CI, sem tratá-los como capacidade medida da VPS.
6. Testes de segurança de métricas: autenticação, ausência de cache/dados sensíveis, erro de banco genérico, cardinalidade e contadores após rollback/retenção. Regressão da suíte SSRF/TLS/SMTP/webhook existente.
7. Roteiro de homologação com inventário somente leitura, domínio, isolamento, recursos, flags desabilitadas, backup/restauração em banco descartável, evidências de aceite e rollback por imagem. CI exercita dump/restore descartável; nenhum comando destrutivo direcionado a servidor real.
8. Revisão de responsividade com Chromium em viewport móvel e preferência de cores, sem envio real. Atualizar threat model, limites do MVP, checklist de release e documentação operacional.

## Limites de operação

Não integrar PRs, configurar VPS/Coolify/Traefik/DNS ou habilitar scheduler/retention/delivery neste incremento. Homologação real depende de inventário/acesso e destinos autorizados. Retenção de eventos/entregas e rotação assistida de chave permanecem requisitos operacionais a decidir antes de uso prolongado; não introduzir exclusões por inferência de autorização da retenção de checks.

## Aceite

- Planejamento, ADR, relatório de PRs e evidências registrados.
- Unitários, SQL com race, benchmark/smoke, backup/restore e frontend/E2E aprovados.
- CI de imagens/dependências verde. Documentar todas as limitações e pendências; não declarar MVP estável nem produção pronta sem homologação.
