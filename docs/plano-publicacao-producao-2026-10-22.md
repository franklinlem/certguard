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

## Situação na primeira inspeção

Na primeira inspeção, produção ainda não havia sido implantada. O acesso à API do Coolify não estava disponível; foi solicitado somente o caminho de arquivo protegido com token. Esse impedimento foi resolvido posteriormente pelo uso dos comandos nativos do Coolify via SSH. Nenhuma mudança em DNS, proxy global, infraestrutura compartilhada ou dados foi realizada. Nenhum segredo foi lido ou versionado.

## Continuação autorizada

Recurso criado pelo responsável: `oiiuvtwkobgvmjeo9j1odxcp`, nome `certguard-producao`, ambiente 8. Consulta pelo bootstrap Laravel do Coolify via SSH confirmou auto-deploy desligado, arquivo de produção correto, porém modo Raw desligado e senha inicial não alfanumérica. Ajustar exclusivamente esse recurso antes da primeira implantação: ativar Raw, substituir senha inicial por aleatória alfanumérica sem exibição e fixar revisão validada. Acesso pelo SSH aos comandos nativos do Coolify elimina a necessidade de provisionar token de API.

Backup desta rodada: dump custom de homologação, criptografado em trânsito pela estação de trabalho, armazenado em diretório exclusivo protegido no frank-note2. Chave de criptografia em arquivo local protegido, separado do servidor de backup; não versionar. Restauração em container PostgreSQL novo, sem rede, sem portas, com dados temporários e limites de CPU/RAM. Conferir schema e contagens sem exibir dados pessoais. Não reutilizar banco real nem remover volumes existentes.


## Resultado

Produção implantada pelo Coolify e HTTPS/readiness/isolamento validados em 22/10/2026. Consulte [evidências finais](validation-producao-2026-10-22.md). Bootstrap administrativo e automação operacional permanecem explicitamente pendentes.
