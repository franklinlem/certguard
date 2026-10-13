# Validação do ajuste de senhas — 13/10/2026

- Testes Go com detector de concorrência: aprovados em todos os pacotes.
- `go vet ./...` e `go build ./...`: aprovados. Cache de compilação colocado em diretório gravável do ambiente local.
- Cobertura específica: mínimo de 9/10 caracteres ASCII e Unicode; símbolos permitidos; limite interno de bcrypt sem truncamento; UTF-8 inválido; senha lida por stdin, LF/CRLF no limite; entrada excessiva limitada; falha de leitura sem expor informações do leitor; hash e login de senha com dez caracteres.
- Verificação de senha existente mantida sem alteração; fixture existente com símbolos continua validada.
- Testes dependentes de PostgreSQL real não executados localmente por falta de banco descartável; permanecem sob responsabilidade da CI.
- Nenhuma migração ou alteração de administrador existente. Aplicação em homologação não atualizada neste ajuste. Produção não alterada.

Decisão: mínimo de dez pontos de código Unicode. Símbolos e espaços continuam permitidos; não se exige composição específica. O limite interno do bcrypt permanece para compatibilidade, com mensagem de senha longa em linguagem comum. Não se promete um máximo uniforme em caracteres para alfabetos diferentes. A leitura remove somente a quebra final de transporte, preservando os demais espaços.
