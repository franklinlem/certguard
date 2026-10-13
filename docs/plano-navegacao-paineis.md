# Navegação para formulário e histórico — 13/10/2026

## Problema e objetivo

Os botões abrem conteúdo distante da posição atual da tela, sem indicar visualmente onde ele está. Após adicionar/editar um endpoint, posicionar o formulário na tela e focar o campo Nome. Após abrir detalhes, posicionar o histórico na tela e focar seu título acessível.

## Implementação e decisões

Manter os painéis existentes e seus dados. Usar referências DOM e uma solicitação explícita de navegação após a renderização; repetir a navegação mesmo quando o painel já estiver aberto. Atualizações de dados e paginação não devem roubar o foco. Expor a relação entre botão e painel com aria-controls/aria-expanded. Usar rolagem imediata, sem animação ou dependência de temporizadores, respeitando pessoas sensíveis a movimento. Manter o título do histórico fora da sequência normal de Tab.

## Aceite e validação

- Adicionar e Editar levam ao formulário e ao campo Nome.
- Detalhes e histórico leva ao título do endpoint escolhido, inclusive em cliques repetidos e trocas de endpoint.
- Recarregar dados não muda o foco automaticamente.
- Testes de regressão de foco e solicitação de rolagem, mais verificações de foco e viewport no E2E desktop/mobile existente.
- Executar testes, lint e build do frontend. Registrar limites de validação local. Entregar em commit e PR quando disponível; sem alterações na VPS ou em produção neste incremento.
