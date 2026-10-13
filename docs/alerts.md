# Alertas e eventos

Novas verificações manuais e automáticas geram eventos quando detectam uma condição de atenção, crítica ou falha técnica. A avaliação acompanha a transação do check; repetições do mesmo problema são deduplicadas. Não há backfill dos checks antigos.

| Situação | Evento / comportamento |
|---|---|
| Faixa de atenção | Vencimento próximo, limiar warning do endpoint |
| Faixa crítica | Vencimento crítico, limiar critical do endpoint |
| Até três dias em faixa crítica | Lembrete distinto se critical >= 3 |
| Expirado / futuro / hostname / cadeia | Problema crítico com causa explícita |
| DNS, conexão, TLS, SSRF | Problema crítico técnico; não afirma certificado expirado |
| Mesmo problema/certificado/faixa | Preserva episódio ativo, sem evento duplicado |
| Condição ou certificado mudou | Resolve anterior e abre novo episódio se necessário |
| Saúde e confiança confirmadas | Resolve problema e registra recuperação informativa |
| Pausa | Preserva problema; não confirma recuperação |
| Troca de destino / arquivamento | Resolve por motivo administrativo, sem recuperação TLS |

No painel, **Problemas ativos** é o filtro inicial. **Todos os eventos** e **Resolvidos e recuperações** mostram o histórico. Atualização explícita consulta um novo snapshot. **Ver observação que originou o evento** mantém os dados do certificado mesmo quando o check original for removido pela retenção.

Esses registros descrevem observações passadas; confira as datas. Eventos não são entregas de notificações. SMTP/webhook, canais cifrados, assinatura e retries estão implementados na etapa 06 e desabilitados por padrão. Consulte [operação](notifications.md). Nenhuma mensagem real foi enviada durante o desenvolvimento.

API: `GET /api/v1/alerts?state=active&limit=20&offset=0`. `state` aceita all, active ou resolved; `endpoint_id` opcional deve ser UUID. Sessão obrigatória. Eventos arquivados continuam consultáveis. A API não oferece exclusão de eventos, resolução manual ou envio de teste.

Não há limpeza automática de eventos ainda. Revisar volume/backup antes de uso prolongado; o próximo planejamento de hardening definirá retenção e índices de acordo com o volume observado. Não modificar serviços compartilhados da VPS para habilitar esta funcionalidade.
