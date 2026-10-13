# ADR-012 — Backup externo cifrado por chave pública

Status: implementado para validação; execução real pendente.

## Decisão

Backup por pull no frank-note2, usando acesso SSH já existente e destino exclusivo no HD /srv/storage. Conferir montagem e UUID; falhar se o disco não estiver disponível. Transmitir pg_dump custom por SSH e cifrar em streaming com OpenPGP antes de persistir qualquer dump.

Chave protegida por senha escolhida pelo responsável, em keyring exclusivo no disco do sistema. Chave pública para execução automática; chave privada e senha para recuperar. Preservar cópia protegida de recuperação em armazenamento separado. Não criar outra identidade SSH, ler chaves privadas, receber senha ou modificar backups anteriores.

Ensaio em PostgreSQL 17 descartável no servidor físico, com rede none, sem portas/volumes existentes e limite de memória. Validar schema e contagens sem revelar dados. Remover somente o container criado pelo ensaio. Não restaurar sobre homologação ou produção.

## Consequências

Backups dependem de conectividade, disco montado e chave pública válida. Recuperação depende da guarda da chave privada/senha. Retenção inicial sem exclusão automática; revisão após 14 dias. Falha deixa backups anteriores intactos. Restauração real é gate separado: testes de script ou checksum não o substituem.
