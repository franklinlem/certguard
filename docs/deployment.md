# Implantação planejada — Coolify e Traefik

**Nenhuma implantação foi executada por esta entrega. Não aplicar configurações globais do proxy nem alterar serviços existentes.**

| Ambiente | Domínio | Projeto Compose | Banco/volume/segredos |
|---|---|---|---|
| Homologação | certguard-dev.franklem.uk | certguard-dev | Exclusivos de homologação |
| Produção | certguard.franklem.uk | certguard-prod | Exclusivos de produção |

## Preparação
1. Fazer inventário somente leitura da VPS: recursos, redes, portas, backups, versão Coolify e proxy Traefik. Confirmar destino DNS e eventual Cloudflare.
2. Criar aplicações/projetos separados no Coolify, usando este repositório privado e `deploy/compose.yaml`; não usar compose.dev na VPS.
3. Selecionar commit explícito. Desabilitar auto-deploy. Definir secrets distintos; senha PostgreSQL alfanumérica forte para a URL Compose.
4. Usar integração nativa do Coolify para conectar somente `web:8080` ao Traefik existente e associar o domínio do ambiente. O nome da rede/entrypoint/certresolver será obtido do inventário; não presumir nomes nem criar segundo Traefik.
5. Não expor API/db diretamente, não montar docker.sock, não compartilhar volumes. O tráfego `/api/` no web alcança somente a API do próprio projeto.

Qualquer mudança de firewall, DNS de serviços existentes, redes compartilhadas, proxy global, armazenamento ou limpeza exige revisão do impacto e autorização específica quando houver risco de interrupção/perda de dados.

## Gate de homologação
- CI verde, build das imagens, migrations, `/health` e `/ready`.
- Certificado HTTPS, isolamento dos ambientes, recursos e logs.
- Testes do fluxo de monitoramento e autenticação quando implementados; esta etapa isolada não habilita o MVP.
- Backup/restore em banco descartável e rollback da aplicação para commit anterior.
- Registrar evidências antes de propor produção. Não habilitar deploy de produção até validação de homologação.

## Backup e rollback
Use pg_dump em formato custom do banco específico. Armazene backup criptografado fora da VPS; secrets e futura encryption key em armazenamento separado. Teste pg_restore em banco descartável e confira migrações, contagens e integridade antes de qualquer restauração real.

Rollback inicial: selecionar imagem/commit anterior; não executar down -v, prune de volumes, DROP DATABASE ou rollback de schema automático. Migração inicial é aditiva. Mudanças destrutivas futuras precisam de plano e autorização.

## Agendamento e retenção

A etapa 03 adiciona a migração 003, que inicializa agendas sem apagar histórico. Na homologação, comece com ambas as tarefas desabilitadas. Cadastre um destino autorizado, habilite o agendador e confirme persistência, cadência e pausa. Avalie o pool adicional de `CHECK_WORKERS + 1` conexões e recursos da VPS. Retenção permanece desabilitada até revisar janela, backup e autorização para remoção definitiva. Consulte [operação](scheduler.md). Não reinicie serviços compartilhados nem altere o Traefik global.
