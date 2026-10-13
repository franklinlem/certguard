# Etapa 04 — dashboard e detalhes TLS

## Objetivo

Entregar a visão geral prevista na v0.4.0: totais, classificação do último check, vencimentos, detalhes técnicos e gráfico de dias restantes. Base: etapa 03, PR #7; nova branch e PR dependente, sem merge automático ou deploy.

## Escopo e decisões

1. API autenticada `GET /api/v1/dashboard`, paginada com os mesmos limites existentes. Agregados sobre todos os endpoints não arquivados, independentemente da página; total, ativos, pausados, saudáveis, atenção, críticos, erros, sem resultado e vencendo entre agora e 30 dias. Contadores de status refletem o último resultado persistido, incluindo pausados, e não garantem saúde atual.
2. Último resultado escolhido por data e UUID para desempate, apenas do hostname/porta atuais. Alterar destino não apresenta o certificado anterior como atual. Histórico anterior permanece acessível e identifica o destino de cada amostra. Snapshot consistente entre agregados e página em transação de leitura.
3. Cards responsivos, lista com status textual, vencimento, último/próximo check e atualização explícita. Estados de carregamento, vazio e erro distintos; nenhum número fictício durante erro ou carregamento.
4. Detalhes do último resultado: destino, IP, porta, emissor, subject, SANs, serial, fingerprint, validade, TLS/cipher e validações. Valores ausentes aparecem como não disponíveis, nunca como falso resultado de validação.
5. Gráfico SVG sem nova dependência: até 20 checks da página de histórico, eixo temporal real e dias restantes no momento do check. Lacunas para falhas sem certificado; tabela textual acessível com todas as amostras. Identificar janela exibida e evitar confundir dias observados com contagem atual. Sem interpolar falhas ou misturar mudanças de hostname.
6. Manter autenticação, CSRF, SSRF, noindex, cache seguro, controles por teclado e espaço reservado para gráfico/carregamento. Eventos e entregas de alertas ficam na etapa seguinte.

## Validação

- PostgreSQL: agregados globais com paginação, último check determinístico, pausados/arquivados, domínio alterado, ausência de certificado, vencimento e banco vazio.
- HTTP: autenticação e validação da paginação da nova rota.
- Frontend: totais/erros, detalhes completos, validação desconhecida e gráfico com lacunas; regressão de CRUD/check.
- E2E: dashboard vazio, cadastro sem resultado, atualização e detalhes sem histórico; nenhum destino externo acessado.
- Tipos, lint, Go race/vet/staticcheck, build e CI de imagens/segurança. Registrar evidências reais.

## Limites e referências

Sem polling, filtros avançados, exportação ou agregação diária nesta etapa. Consultas aos últimos resultados são indexadas por endpoint/data; reavaliar desempenho antes de grandes volumes. Não modificar VPS nem habilitar agendador/retenção.

Referências de interface: [Search Central](https://developers.google.com/search/docs?hl=pt-br), [cache](https://developer.chrome.com/docs/performance/insights/cache?hl=pt-br), [CLS](https://web.dev/articles/optimize-cls?hl=pt-br). Painel privado permanece sem indexação, API sem cache, SVG com proporção definida.
