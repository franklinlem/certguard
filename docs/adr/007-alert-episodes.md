# ADR-007 — episódios de alerta transacionais

Status: proposto na etapa 05; depende do ADR-006.

Primeiro validamos a geração e o ciclo de vida dos eventos; a entrega externa será um incremento separado com canais, outbox, retries e proteção SSRF de webhooks. Não há notificações enviadas nesta etapa, nem necessidade de novos segredos.

Cada check novo avalia regras dentro de sua transação, após persistir o resultado e reagendar. O lock do endpoint já adquirido por checks manuais/automáticos serializa a avaliação. Um índice único parcial impede dois problemas ativos para o mesmo endpoint. Falha ao salvar o evento reverte check, agenda e resolução do evento anterior.

A chave inclui hostname, porta, tipo da regra, severidade, limiar e fingerprint quando disponível. É usada para comparar com o episódio ativo, sem unicidade histórica: uma recorrência após recuperação deve abrir outro episódio. Repetições de dias restantes dentro da mesma faixa não criam eventos. Um certificado novo ainda problemático abre outro episódio. Não há repetição por tempo nesta entrega.

Faixas warning/critical usam a classificação e limiares do check. Até três dias na faixa crítica cria um lembrete distinto quando o limiar crítico é de pelo menos três dias. Limiares críticos menores continuam sua própria regra. Expirado, certificado futuro, hostname incompatível, cadeia inválida e falhas de inspeção são problemas críticos identificados por tipo; uma falha de DNS não é descrita como expiração. Recuperação exige resultado saudável, código success e hostname/cadeia confirmados. Resultado incompleto não confirma recuperação.

Transições resolvem o anterior com `changed`; saúde confirmada usa `recovered` e gera evento informativo já resolvido. Troca de hostname e arquivo resolvem com `target_changed`/`archived` dentro da transação do endpoint, sem evento de recuperação. Pausa mantém o problema histórico ativo. Alteração de nome não recria o episódio. Limiares novos são avaliados no próximo check.

Evento guarda snapshot do destino, nome e resultado. Referência ao check permite rastreio, mas `ON DELETE SET NULL` impede que a retenção apague eventos ou fique bloqueada. A migração cria somente estrutura/índices, sem reprocessar checks antigos. Eventos não têm limpeza automática nesta etapa; retenção de 365 dias do planejamento global será considerada no hardening, com backup e autorização conforme risco.

Consultas autenticadas usam snapshot read-only consistente entre contadores e página. Filtro de estado afeta a lista; indicadores incluem todos os estados (ou apenas o endpoint solicitado). `resolved` inclui problemas resolvidos e recuperações; `summary.resolved` conta só problemas resolvidos, e `summary.info` conta recuperações. Não equivale a confirmação de entrega de mensagens.
