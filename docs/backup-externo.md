# Backup externo e ensaio de restauração

## Locais exclusivos

- Servidor físico: frank-note2, usuário franklin.
- HD: /srv/storage, identidade conferida por UUID antes de escrever. /mnt/sda1 é outra montagem do mesmo HD; não constitui segunda cópia.
- Arquivos cifrados: /srv/storage/backups/certguard/{homologacao,producao}, modo 0700.
- Scripts: ~/.local/lib/certguard-backup, fora das rotinas existentes.
- Chave OpenPGP protegida e configuração: ~/.local/share/certguard-backup. Não versionar essa pasta. A senha de recuperação é entrada somente pelo responsável.
- Logs e lock: ~/.local/state/certguard-backup. Não contém dumps ou senhas.

## Fluxo

O script usa a identidade SSH já existente do backup Contabo, sem copiar ou ler a chave privada. A fonte executa apenas pg_dump custom do container CertGuard específico. SSH transporta o dump até GnuPG, que cifra em streaming para o HD. Nenhum dump aberto é gravado em disco. Um arquivo parcial exclusivo só é publicado após sucesso da leitura e cifragem; checksum SHA256 é armazenado ao lado. Execuções concorrentes recusadas, nomes únicos, sem sobrescrita ou limpeza de backups anteriores.

`backup.sh homologacao` é a fonte já conhecida. `backup.sh producao` exige o arquivo `production-resource` com o UUID do recurso de produção; recusa o UUID de homologação. Não configurar antes de provisionar e validar o recurso correto.

`daily.sh` executa os ambientes configurados em sequência. Agendamento proposto: 05:30 no fuso do servidor, posterior ao backup Contabo das 04:30. Preservar entradas cron anteriores. Inicialmente conservar todos os backups; revisar retenção após 14 dias. Nenhuma exclusão automática de arquivos antigos nesta entrega.

A variável opcional CERTGUARD_BACKUP_MOUNT permite testar os scripts com um disco simulado; mesmo assim o UUID deve coincidir com storage.uuid. O agendamento real usa /srv/storage, sem override.

## Chave de recuperação

O responsável cria a chave em keyring exclusivo, com senha digitada no terminal. A chave pública é suficiente para backups automáticos; recuperar exige chave privada e senha. Guardar uma cópia protegida da chave privada em armazenamento separado, além da senha no gerenciador. Não receber, imprimir, enviar em chat ou versionar senha/chave privada. Não basta guardar somente a chave pública. Referência: [gerenciamento de chaves GnuPG](https://gnupg.org/documentation/manuals/gnupg/OpenPGP-Key-Management.html).

## Ensaio seguro

`restore-test.sh /srv/storage/backups/certguard/homologacao/ARQUIVO.dump.gpg` verifica caminho, montagem, UUID e checksum. Usa somente PostgreSQL descartável com rede none, sem portas e sem volumes de aplicações. Memória limitada a 512 MiB, dados em tmpfs; imagem PostgreSQL 17. A chave precisa estar desbloqueada no GnuPG pelo responsável. Em execução automatizada o script falha sem solicitar senha, em vez de registrar ou receber segredo.

pg_restore usa --exit-on-error e banco novo. Consultas finais mostram somente contagens de tabelas e versão de schema. O container criado pelo ensaio é removido ao terminar; backups e demais containers permanecem. Para dumps que excedam a capacidade do tmpfs, planejar outro ensaio isolado; nunca improvisar restauração sobre banco existente.

Sucesso de cifragem/checksum não comprova restaurabilidade. Somente registrar o gate como aprovado após ensaio completo com uma cópia real. Não restaurar dados/sessões de homologação em produção.

## Falhas e recuperação

Disco ausente/incorreto, espaço baixo, chave indisponível/expirada, erro SSH e falha de cifragem interrompem o backup. Arquivo parcial próprio é removido, backups anteriores preservados. Conferir log, conectividade e montagem sem alterar outros serviços. Não contornar verificação SSH, TLS ou identidade de disco. A chave OpenPGP tem validade: renovar sob supervisão antes do vencimento e preservar capacidade de recuperar os backups antigos.
