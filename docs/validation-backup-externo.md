# Backup externo — validação real (13/10/2026)

## Resultado comprovado

- Acesso pelo Remote Desktop Commander ao frank-note2 como franklin. HD /dev/sda1 ext4 montado em /srv/storage, 433 GB livres; UUID validado pelos scripts antes de escrever.
- Chave OpenPGP CertGuard Backup criada pelo responsável em keyring exclusivo, com entrada de senha somente no terminal. Fingerprint público: 4D6D40444014667AC0370D3AF7FCB1EC02857CD4. Senha e chave privada não recebidas, copiadas ou versionadas pelo agente.
- Revisão final dos três scripts do PR #17 instalada em ~/.local/lib/certguard-backup com modo 0700. Configuração e backups com modo restrito.
- Primeiro backup real da homologação: certguard-20261013T213526Z-YwHmqR.dump.gpg, 6396 bytes; checksum SHA256 aprovado. Dump custom transportado por SSH e cifrado em streaming.
- Imagem oficial postgres:17-alpine baixada sem alterar DNS; digest sha256:b0f9560a2de083e2cc7382e75f808c7381a32852a7ec49117deedb300e552b24.
- Após desbloqueio da chave pelo responsável, restauração real aprovada em PostgreSQL descartável, rede none, sem portas/volumes de aplicações e com dados em tmpfs. Resultado: migrations=6, schema=6, users=1, endpoints=3, checks=4, events=0, channels=0, deliveries=0.
- Container do ensaio removido; comparação de IDs/nomes dos cinco containers preexistentes sem diferenças. Nenhuma restauração sobre banco real.
- Cron exclusivo instalado para 05:30 diariamente, fuso local America/Sao_Paulo (-03). Conteúdo anterior preservado e cópia anterior em ~/.local/state/certguard-backup. Serviço cron ativo. Cron não altera timers Contabo/Hermes.
- Execução manual da mesma rotina daily.sh aprovada: segunda cópia certguard-20261013T213817Z-EciFkR.dump.gpg. Produção ignorada porque production-resource não está configurado. Nenhuma exclusão automática de backups.
- Oito testes dos scripts executados novamente e aprovados; bash -n e git diff --check aprovados. CI Validação e Segurança do commit d31f45612961e26547e999b905a0f2296c577f5a aprovadas antes desta atualização documental.

## Limites e próximos passos

- Ainda não ocorreu a primeira execução pelo cron: prevista para 14/10/2026 às 05:30. Execução manual da rotina comprovada; log automático em ~/.local/state/certguard-backup/daily.log.
- Responsável deve guardar cópia protegida da chave privada e certificado de revogação em armazenamento separado deste computador, além da senha no gerenciador. Isso não foi realizado nem registrado como aprovado; guardar apenas a chave pública não permite recuperar backups.
- Backups dependem de frank-note2 ligado, HD montado e conexão com a VPS. A rotina falha se disco/UUID/acesso/chave pública não estiverem disponíveis; não há notificação externa de falha configurada nesta etapa.
- Produção não implantada. Backup de produção só será configurado após identificar o recurso exclusivo correto. Agendamento TLS e envio de alertas mantêm gates próprios.
