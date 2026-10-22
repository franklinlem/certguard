# Publicação de produção — 22/10/2026

## Solicitação e situação confirmada

Franklin solicitou concluir o ambiente de produção, mantendo a homologação existente. O Compose de produção já está versionado em main: `compose.coolify.prod.yaml`. Não foi necessário alterar aplicação ou criar outro Compose.

Verificações desta sessão, somente leitura:

- SSH existente para Contabo confirmou hostname `vmi3570746`.
- Homologação respondeu HTTPS 200 com validação de certificado; db, API e web estão saudáveis. O comando `certguard healthcheck` terminou com código zero.
- HTTPS de produção falhou por certificado autoassinado; não se utilizou bypass TLS. Isso não comprova sozinho ausência de recurso no painel.
- Listagem de containers em execução mostrou somente o recurso CertGuard de homologação `rk48swz6mdqduoyyhfnqhylp`.
- Memória disponível: 5250 MiB; disco raiz: 73 GiB livres. Medições pontuais, sem teste de carga.
- Web de homologação participa somente das redes `coolify` e do próprio recurso na inspeção realizada.
- SSH existente para `frank-note2` confirmou o hostname; backup e restauração ainda não executados nesta sessão.
- Acesso ao diretório administrativo do Coolify foi recusado; sudo não interativo exige senha. Nenhuma permissão ou mecanismo de acesso foi alterado.

## Plano de execução

1. Obter acesso à API do Coolify por token disponibilizado em arquivo protegido, sem exibi-lo em logs/conversa ou incluí-lo no Git. Inventariar projeto, ambientes, recurso e revisão implantada; verificar CI da revisão escolhida.
2. Verificar backup externo específico do CertGuard e restauração em banco descartável, usando o acesso existente ao frank-note2. Definir destino e criptografia conforme estrutura existente, sem alterar outros backups.
3. Criar recurso separado no ambiente de produção, modo Compose Raw, arquivo `/compose.coolify.prod.yaml`, revisão explícita e auto-deploy desligado. Gerar segredo exclusivo no recurso; não copiar banco, sessões ou administrador de homologação.
4. Implantar preservando serviços existentes. Verificar migração com saída zero, saúde dos serviços, redes/volume próprios, ausência de portas públicas de banco/API e HTTPS confiável em `certguard.franklem.uk`.
5. Realizar bootstrap administrativo por entrada segura e validar autenticação/readiness. Scheduler, retenção, entregas e métricas permanecem desligados conforme configuração inicial, até homologação específica.
6. Registrar UUID, SHA, evidências e eventual rollback da aplicação, preservando volume/schema. Atualizar documentação somente com resultados efetivamente comprovados.

## Limite atual

Produção não foi implantada nesta sessão. Falta acesso autenticado à API do Coolify; foi solicitado somente o caminho do arquivo protegido com token. Nenhuma mudança em DNS, proxy global, infraestrutura compartilhada ou dados foi realizada. Nenhum segredo foi lido ou versionado.
