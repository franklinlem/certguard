# Produção — evidências de 22/10/2026

## Preparação e backup

Recurso Coolify `oiiuvtwkobgvmjeo9j1odxcp`, projeto CertGuard, ambiente `producao`, nome `certguard-producao`. Arquivo `/compose.coolify.prod.yaml`, base `/`, Compose Raw ativado e auto-deploy desativado. Senha PostgreSQL inicial substituída por valor alfanumérico aleatório exclusivo, sem exibição ou versionamento. Revisão fixada em `93d1f5af0e26f588c2e21b8f039c4f5f4ff9c540`, a mesma da homologação. Checks `backend`, `frontend`, `docker`, `dependencies` e `source-and-images` aprovados no GitHub nessa revisão.

Docker Compose instalado na Contabo renderizou o arquivo com identidade real de produção e senha fictícia para validação; `scripts/check-coolify-compose.py oiiuvtwkobgvmjeo9j1odxcp producao` aprovou isolamento e roteamento.

Backup custom de homologação criado em 22/10/2026 às 13:19:36 -03, criptografado com GPG/AES256 antes de armazenamento fora da VPS:

- Arquivo no frank-note2: `/home/franklin/backups/certguard/homologacao/20261022T161936Z.dump.gpg`.
- Permissão 0600, tamanho 24826 bytes, SHA256 `7fe033f7620d5745433eba1738b33aaab213a4447fc29c9c60d9a8955a71649a`.
- Chave separada na estação: `/home/franklin/.local/share/certguard/backup-encryption.key`, permissão 0600. Preservar em armazenamento seguro; sem ela o backup não pode ser recuperado. A chave não foi copiada para a VPS ou servidor de backup.
- Nenhum dump em texto claro foi gravado em disco; transferência por SSH e descriptografia em stream.

Restauração executada no frank-note2 em container novo `certguard-restore-20261022t161936z`, PostgreSQL 17, rede `none`, sem portas publicadas, dados em tmpfs, RAM 512 MiB e 1 CPU. `pg_restore --exit-on-error` aprovado. Comparação antes/depois da origem e da restauração: seis migrações, versão máxima 6; 1 usuário, 1 sessão, 3 endpoints, 4 checks, zero eventos/canais/entregas e métricas duráveis equivalentes. Nenhum dado pessoal foi exibido. Container de ensaio parado após validação; nenhum volume existente foi removido.

Este é um backup pontual validado. Agendamento de backup e política de retenção ainda não foram configurados. As entregas estão desligadas e não há chave de canais configurada para copiar nesta etapa.

## Implantação

Deploy manual enfileirado pelo mecanismo nativo do Coolify via SSH, UUID `01m5j6n193zq38zhenghf18zt5`. Concluído com status `finished`; API iniciou às 13:23:07 -03. Migração encerrou com código zero, schema com seis migrações e versão máxima 6. Nenhuma alteração global de proxy, DNS, firewall ou outros projetos realizada.


## Verificações após implantação

- HTTPS com validação de certificado aprovada, sem bypass: `/`, `/api/health` e `/api/ready` retornaram 200 em produção e homologação.
- `/api/v1/endpoints` sem sessão retornou 401 nos dois ambientes; `/metrics` retornou 404 na interface pública.
- db, API e web de produção saudáveis; migrate encerrado com código 0.
- Banco somente na rede `oiiuvtwkobgvmjeo9j1odxcp_database`, API nas redes database/app, web nas redes app/coolify. Volume exclusivo `oiiuvtwkobgvmjeo9j1odxcp_postgres_data`. Nenhum serviço do recurso publica portas no host.
- API/web/migrate com filesystem somente leitura; limites efetivos: banco 512 MiB, API/migrate 256 MiB, web 128 MiB.
- Origem efetiva `https://certguard.franklem.uk`, `APP_ENV=production`; scheduler, retenção e entregas continuam desligados. Monitoramento automático não está ativo.
- Homologação continuou saudável; listagem pós-deploy manteve demais containers em execução, com uptimes anteriores. Isso não mede disponibilidade contínua de todos os serviços.
- Produção contém zero usuários; bootstrap administrativo ainda depende da senha escolhida pelo responsável. Login/logout autenticado de produção não foram ensaiados. Procedimento em [bootstrap-producao.md](bootstrap-producao.md).

## Pendências operacionais

Criar administrador por entrada segura, validar login/logout, definir rotina recorrente de backup e retenção, proteger cópia de recuperação da chave local e homologar scheduler/canais antes de ativar monitoramento automático. O backup validado nesta rodada é da homologação; o banco novo de produção não recebeu dados ou usuários copiados. Não apresentar a implantação inicial como aceite integral do MVP.
