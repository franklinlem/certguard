# Backup externo CertGuard — 13/10/2026

## Inventário confirmado

Servidor frank-note2, usuário franklin com Docker. Disco /dev/sda1 ext4 já montado em /srv/storage, 433G disponíveis conforme informação do responsável. /mnt/sda1 é segunda montagem do mesmo disco, não um destino independente. Backups Contabo e Hermes existentes, preservados. Serviço backup-contabo usa SSH franklin@158.220.120.2 com identidade existente /home/franklin/.ssh/contabo_backup; acesso ao Docker da VPS confirmado sem sudo. Não ler/copiar a chave privada SSH.

## Decisão e execução

- Diretório exclusivo /srv/storage/backups/certguard, modo 0700, ambiente homologacao separado de producao.
- Chave OpenPGP exclusiva, protegida por senha escolhida pelo responsável, em ~/.local/share/certguard-backup/gnupg no disco do sistema. Gerar por entrada interativa do responsável; não receber senha na conversa. Backup usa somente a chave pública. Preparar cópia de recuperação protegida para armazenamento separado, sem versionar segredos.
- Script Bash com validação de mountpoint, fonte de disco, espaço mínimo, identidade do servidor, nome de container e ambientes permitidos. pg_dump custom transmitido por SSH e cifrado em streaming, sem dump aberto em disco. Arquivo parcial exclusivo, publicação somente após sucesso de todas as etapas e checksum do arquivo cifrado. Nenhuma sobrescrita, nenhuma exclusão de backups anteriores.
- Reutilizar a identidade SSH existente. Não alterar scripts, cron/timers ou diretórios dos outros backups. Bloquear execuções concorrentes com flock.
- Agendamento separado diário às 05:30, após rotina Contabo das 04:30; preservar integralmente entradas cron existentes. Inicialmente conservar todos os arquivos. Revisar retenção após 14 dias; não habilitar exclusão automática sem autorização específica.
- Restaurar uma cópia em PostgreSQL 17 descartável no frank-note2, rede none, sem porta ou volume de aplicações existentes, memória limitada. Usar container nomeado exclusivo e tmpfs; verificar schema, administrador, endpoints e histórico sem imprimir linhas sensíveis. Remover somente o container de teste criado pelo procedimento, nunca bancos/volumes existentes.
- Verificar bash, casos de falha (disco ausente, ambiente inválido, falha SSH/cifragem), backup real, checksum e restauração antes de registrar aprovação. Não ativar produção antes de concluir os gates de homologação.

## Dependências de interação

A senha da chave OpenPGP deve ser digitada pelo responsável no próprio terminal. O agente pode cifrar usando chave pública; para o ensaio de restauração, o responsável desbloqueia a chave pelo GnuPG sem compartilhar senha. Se não houver cache de desbloqueio, pausar somente a restauração e solicitar essa entrada segura.
