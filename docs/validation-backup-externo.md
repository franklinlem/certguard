# Backup externo — validação em andamento (13/10/2026)

## Comprovado

- Conexão temporária autenticada pelo responsável permitiu acessar frank-note2 como franklin; nenhuma senha foi recebida.
- /srv/storage corresponde ao HD ext4 /dev/sda1, já montado antes desta etapa. /mnt/sda1 criado pelo responsável é segunda montagem do mesmo disco. Diretório /srv/storage/backups/certguard criado com modo 0700, separado de Contabo/Hermes.
- Rotina backup-contabo existente permanece intacta. Identidade SSH já utilizada por ela permite executar Docker na VPS como franklin. Não criada, lida ou copiada chave privada SSH; não alteradas permissões sudo/SSH.
- Preparadas pastas exclusivas em ~/.local/share/certguard-backup, ~/.local/lib/certguard-backup e ~/.local/state/certguard-backup. UUID do HD registrado no servidor físico, sem versionar configuração específica.
- Scripts de backup e ensaio preparados; primeira cópia instalada ainda requer sincronização com a revisão final antes de executar.
- Oito testes aprovados com chave GPG descartável: cifragem real e decriptação idêntica à fixture; nomes distintos sem sobrescrita; preservação de arquivo anterior; disco ausente/incorreto; ambiente inválido/produção não configurada; falha SSH; falha de cifragem; lock de concorrência; checksum adulterado/caminho externo recusados antes de criar container. Sintaxe Bash e git diff --check aprovados. Nenhuma chave de teste é usada no ambiente real.

## Dependências e limite atual

A primeira tentativa de geração da chave pelo responsável juntou dois comandos e falhou no export. Enviado comando único correto, com entrada segura da senha; criação ainda não confirmada. Nenhum backup real gerado, nenhuma restauração real executada e nenhum agendamento instalado nesta evidência.

Download de postgres:17-alpine no servidor físico falhou por resolução de auth.docker.io. Não alterado DNS global. Imagem existente na VPS identificada como sha256:b0f9560a2de083e2cc7382e75f808c7381a32852a7ec49117deedb300e552b24; transferência para ensaio ainda não concluída. Usar imagem oficial já validada, conferindo ID após importação.

Conexão SSH temporária encerrou; solicitado ao responsável reabrir a sessão, sem criar credenciais persistentes. Hash de IDs/nomes dos containers físicos antes do ensaio: 21c29d4002d21382a822a520fe9769d51132f614e97d57ec0910731351e09616. Nenhum container de restauração iniciado até aqui.

## Próximas verificações

1. Confirmar chave protegida, registrar somente fingerprint público e instalar revisão final dos scripts.
2. Gerar backup real cifrado da homologação e conferir checksum.
3. Desbloqueio seguro da chave pelo responsável; ensaio completo em container PostgreSQL novo, sem rede/portas/volumes existentes, com limpeza exclusiva do próprio container.
4. Guardar cópia protegida de recuperação em armazenamento separado, sem receber senha ou expor chave privada.
5. Instalar agendamento diário separado às 05:30, preservando cron e timers existentes; manter todos os backups, sem exclusão automática nesta fase.
6. Registrar evidências antes do aceite e da implantação de produção. Agendamento TLS e entrega de alertas continuam com seus gates próprios e permanecem desligados.
