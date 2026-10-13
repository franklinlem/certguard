# Validação da etapa 04

Base: etapa 03 / PR #7. Sem implantação ou mudança na VPS.

## Local

- Go: testes com detector de corrida e vet aprovados. PostgreSQL não disponível localmente: integração SQL será confirmada na CI.
- Frontend: tipos, lint, oito testes e build aprovados.
- Inspeção visual local com Playwright/Chromium e dados simulados: desktop, detalhes e mobile de 390px, sem overflow horizontal. Esta prévia não substitui integração com banco.
- Casos: agregados globais, exclusão de arquivados, pausados, troca de destino, desempate do último resultado, erro após sucesso, gráfico com falhas e datas reais, dados ausentes e estado de indisponibilidade.

## CI

Aguardando execução para confirmar PostgreSQL, imagens, E2E e segurança. Não apresentar a ausência de banco local como integração aprovada.

## Limites

Status reflete o último check persistido; a interface mostra quando foi observado. Gráfico limitado aos 20 checks da página; tabela acessível acompanha as amostras. Alertas e entregas continuam fora desta etapa. Agendador e retenção não foram habilitados; PRs anteriores continuam dependências.
