# Validação da navegação para painéis — 13/10/2026

## Resultado

- Testes de componentes: 15 aprovados em quatro arquivos.
- ESLint, TypeScript e build Vite: aprovados.
- Playwright: dois cenários aprovados, desktop (1366×900) e celular (390×844). Usado frontend local com API simulada, sem acesso ao servidor ou aos endpoints externos.
- Com vinte endpoints, abrir formulário posiciona campo Nome e título no viewport; clicar novamente preserva o rascunho e repõe o foco. Histórico recebe foco e fica visível enquanto a resposta de dados permanece pendente. Reabrir o mesmo histórico e trocar para outro endpoint funciona. Editar o último endpoint retorna ao formulário e campo Nome.
- Teste de componente confirma que atualizar os dados não provoca nova rolagem nem rouba o foco.
- O E2E administrativo existente recebeu verificações de foco/viewport, mas o fluxo completo com PostgreSQL não foi executado localmente; permanece para CI.

## Decisão e limites

Usar navegação explícita após a renderização, com refs DOM e rolagem imediata. O título do histórico tem tabindex -1; o campo Nome mantém a navegação normal de teclado. Nenhuma mudança em autenticação, API ou banco neste incremento. A correção anterior de senhas permanece no histórico local e aguarda publicação junto desta entrega. Homologação e produção não atualizadas nesta validação.
