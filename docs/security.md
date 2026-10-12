# Segurança

A política SSRF e o fluxo de confiança estão nos ADR-002 e ADR-003. Resultados de certificado são dados não confiáveis; a futura UI deve renderizar texto, nunca HTML recebido do certificado. Secrets não entram no Git, logs ou argumentos da CLI.

O checker não está acessível via HTTP. Não publicar CRUD/checks antes de autenticação, autorização, CSRF, rate limiting e limites de concorrência. O modo atual expõe apenas página informativa e saúde. Banco isolado em rede interna, sem portas públicas. Docker API/web sem root, filesystem somente leitura, capabilities removidas e no-new-privileges.

PostgreSQL usa comportamento padrão da imagem oficial para inicializar volume e reduzir privilégios; não forçar read-only em seu diretório de dados. Senhas são individuais por ambiente. Para compose use senha alfanumérica; credenciais com caracteres reservados exigem codificação de URL.

Cabeçalhos de segurança e noindex na página de desenvolvimento. HSTS deve ser configurado no proxy somente após validar HTTPS e domínio; não aplicar includeSubDomains ao domínio compartilhado automaticamente.

Antes da homologação operacional: rever egress/rotas, proteger main, habilitar controles de segurança disponíveis no plano GitHub, validar restore e revisar segredos. CodeQL em repositório privado depende da disponibilidade do GitHub Code Security; não ativar cobrança automaticamente.
