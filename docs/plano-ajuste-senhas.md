# Ajuste da orientação de senhas

## Objetivo e critérios

O administrador deve receber a orientação “Escolha uma senha com no mínimo 10 caracteres. Você pode usar letras, números e símbolos.” Senhas menores devem produzir uma mensagem específica, sem revelar a senha. Contar caracteres Unicode, em vez do tamanho da representação UTF-8. Não exigir combinações específicas nem limitar senhas a letras e números.

## Implementação

Centralizar a validação no pacote de autenticação, reutilizá-la no bootstrap e validar a entrada limitada da CLI antes de acessar o banco. Preservar bcrypt e seu limite interno de entrada, usando uma mensagem compreensível para entradas longas. Não mudar a verificação de credenciais existentes. Atualizar as orientações de criação segura, sem recriar o administrador.

## Validação e entrega

Testar limites de 9/10 caracteres, Unicode, símbolos, limite de bcrypt, entrada por stdin com LF/CRLF e erros de leitura. Executar testes Go com detector de concorrência, vet e build. Registrar evidências e preparar PR. Nenhuma alteração de infraestrutura ou publicação em produção faz parte deste ajuste. A regra nova só estará disponível na homologação após a atualização da aplicação.
