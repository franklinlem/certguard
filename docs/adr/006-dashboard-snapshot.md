# ADR-006 — indicadores por último resultado e snapshot de leitura

Status: proposto na etapa 04, dependente da etapa 03.

A rota autenticada `/api/v1/dashboard` reúne contadores globais e uma página de endpoints. Duas consultas usam a mesma transação read-only em REPEATABLE READ, evitando que um check concorrente torne os totais e a página incompatíveis. A leitura não dispara inspeções nem altera agenda.

Usamos o último check do hostname/porta atuais, por `checked_at DESC, id DESC`, através de lateral join. O índice existente por endpoint/data apoia essa consulta; o UUID desempata timestamps iguais. Trocar o hostname mantém o histórico, mas não reaproveita como atual o resultado de outro destino. Voltar a um hostname anterior pode recuperar seu último resultado conhecido; a data observada permanece explícita. Não há nova migração.

Contadores incluem todos os endpoints não arquivados, inclusive pausados. Status é a classificação persistida no check, não recalculada a cada visita ou após alterar limiares. “Sem resultado” não significa saudável. Vencendo em 30 dias considera `not_after` do último resultado entre o instante do snapshot e 30 dias seguintes; exclui expirados e resultados sem certificado. Uma falha recente não reutiliza o certificado de um sucesso antigo. A idade do resultado permanece visível; saúde atual requer nova inspeção.

A interface utiliza atualização explícita, mantendo carga previsível. O gráfico representa até 20 registros da página de histórico, ordenados temporalmente, sem agregação nem amostragem oculta. Eixo X proporcional ao tempo, eixo Y em dias observados; falha sem validade e mudança de destino interrompem segmentos. Renovações podem causar saltos legítimos. Uma tabela oferece os mesmos dados para leitura sem depender de cor, ponteiro ou SVG. O histórico pode ser reduzido pela política de retenção.

Não adicionamos biblioteca de gráficos para este caso simples. SVG tem dimensões/proporção definidas; conteúdo recebido de certificados continua renderizado como texto. Noindex e política de cache existentes são preservados. Testes de carga e filtros avançados pertencem a futuros incrementos; revisar os planos de consulta antes de ampliar muito o volume.
