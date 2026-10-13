# Validação da etapa 04

Base: etapa 03 / PR #7. Sem implantação ou mudança na VPS.

## Local

- Go: testes com detector de corrida, vet e staticcheck aprovados. PostgreSQL não disponível localmente; integração SQL aprovada na CI.
- Frontend: tipos, lint, oito testes e build aprovados.
- Inspeção visual local com Playwright/Chromium e dados simulados: desktop, detalhes e mobile de 390px, sem overflow horizontal. Esta prévia não substitui integração com banco.
- Casos: agregados globais, exclusão de arquivados, pausados, troca de destino, desempate do último resultado, erro após sucesso, gráfico com falhas e datas reais, dados ausentes e estado de indisponibilidade.

## CI

Código validado: `2265d61b1a2d6bb3270c0d69ea6a43146e495c9a`.

- [Validação #38327418880](https://github.com/franklinlem/certguard/actions/runs/38327418880): backend com PostgreSQL/race/vet/staticcheck, cobertura, fuzzing, frontend, Docker e Playwright aprovados.
- [Segurança #38327418886](https://github.com/franklinlem/certguard/actions/runs/38327418886): dependências, código e imagens aprovados.
- [PR #8](https://github.com/franklinlem/certguard/pull/8) aberto sobre a branch da etapa 03, sem merge.

A atualização posterior apenas registra evidências e ajusta a documentação OpenAPI (descrição/versão e resultado nullable), sem alterar o código validado. Docker/PostgreSQL/E2E integrado foram executados na CI; a prévia visual local usou dados simulados.

## Limites

Status reflete o último check persistido; a interface mostra quando foi observado. Gráfico limitado aos 20 checks da página; tabela acessível acompanha as amostras. Alertas e entregas continuam fora desta etapa. Agendador e retenção não foram habilitados; PRs anteriores continuam dependências.
