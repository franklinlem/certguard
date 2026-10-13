# Implantação planejada — Coolify e Traefik

**Homologação inicial implantada em 13/10/2026; produção não implantada. Consulte [evidências e pendências](validation-etapa-08.md). Não alterar configurações globais do proxy nem serviços existentes.**

| Ambiente | Domínio | Projeto Compose | Banco/volume/segredos |
|---|---|---|---|
| Homologação | certguard-dev.franklem.uk | certguard-dev | Exclusivos de homologação |
| Produção | certguard.franklem.uk | certguard-prod | Exclusivos de produção |

## Preparação
1. Fazer inventário somente leitura da VPS: recursos, redes, portas, backups, versão Coolify e proxy Traefik. Confirmar destino DNS e eventual Cloudflare.
2. Criar aplicações/projetos separados no Coolify, usando este repositório privado. Na homologação inventariada, usar `/compose.coolify.yaml`, base `/` e modo Raw, conforme [ADR-010](adr/010-coolify-isolation.md); não usar compose.dev na VPS.
3. Selecionar commit explícito. Desabilitar auto-deploy. Definir secrets distintos; senha PostgreSQL alfanumérica forte para a URL Compose.
4. Usar os labels do Compose Raw de homologação para conectar somente `web:8080` ao Traefik existente e associar o domínio do ambiente. O nome da rede/entrypoint/certresolver será obtido do inventário; não presumir nomes nem criar segundo Traefik.
5. Não expor API/db diretamente, não montar docker.sock, não compartilhar volumes. O tráfego `/api/` no web alcança somente a API do próprio projeto.

Qualquer mudança de firewall, DNS de serviços existentes, redes compartilhadas, proxy global, armazenamento ou limpeza exige revisão do impacto e autorização específica quando houver risco de interrupção/perda de dados.

## Gate de homologação
- CI verde, build das imagens, migrations, `/health` e `/ready`.
- Certificado HTTPS, isolamento dos ambientes, recursos e logs.
- Testes do fluxo de monitoramento/autenticação, canais, entregas e métricas; CI isolada não conclui homologação.
- Backup/restore em banco descartável e rollback da aplicação para commit anterior.
- Registrar evidências antes de propor produção. Não habilitar deploy de produção até validação de homologação.

## Backup e rollback
Use pg_dump em formato custom do banco específico. Armazene backup criptografado fora da VPS; secrets e encryption key em armazenamento separado. Teste pg_restore em banco descartável e confira migrações, contagens e integridade antes de qualquer restauração real.

Rollback inicial: selecionar imagem/commit anterior; não executar down -v, prune de volumes, DROP DATABASE ou rollback de schema automático. Migração inicial é aditiva. Mudanças destrutivas futuras precisam de plano e autorização.

## Agendamento e retenção

A etapa 03 adiciona a migração 003, que inicializa agendas sem apagar histórico. Na homologação, comece com ambas as tarefas desabilitadas. Cadastre um destino autorizado, habilite o agendador e confirme persistência, cadência e pausa. Avalie o pool adicional de `CHECK_WORKERS + 2` conexões e recursos da VPS. Retenção permanece desabilitada até revisar janela, backup e autorização para remoção definitiva. Consulte [operação](scheduler.md). Não reinicie serviços compartilhados nem altere o Traefik global.

## Entregas de alertas

Etapa 06 requer migração 005 e chave externa somente para cadastrar/ativar canais. DELIVERY_ENABLED=false permanece padrão. Antes de habilitar em homologação, valide destinatário/receptor autorizado, backup separado da chave, orçamento de conexões (CHECK_WORKERS+2) e backlog. Nunca trocar a chave de instalação existente sem migração. Consulte [operação](notifications.md). Nenhum envio ou alteração da VPS foi realizado nesta entrega.

## Gate consolidado

Seguir [roteiro de homologação](homologacao.md). Fixar SHA e schema esperado (1–6), verificar readiness, revisar backlog/flags, backup separado da chave e compatibilidade do rollback. METRICS_TOKEN vazio mantém endpoint desabilitado; coletor interno é configuração posterior, sem instalação automática. PRs e dependências foram integrados em 13/10/2026; consulte [evidências](integracao-prs-2026-10-13.md) e [inventário](inventario-homologacao-2026-10-13.md) antes da primeira implantação.

## Destino de backup informado

Responsável escolheu franklin@frank-note2 (servidor físico Linux Mint XFCE com Docker, Hermes, outros backups e Uptime Kuma). Verificação adiada por solicitação dele. Não é backup configurado: diretório, transporte, criptografia, retenção e restauração permanecem pendentes. Não acessar ou modificar serviços desse host sem o próximo escopo definido.


## Configuração exclusiva de produção (preparada, não implantada)

Usar `/compose.coolify.prod.yaml` em recurso novo do ambiente `producao`, no modo Raw. O recurso recebe UUID exclusivo, segredo PostgreSQL exclusivo gerado no Coolify e volume próprio. O domínio e a origem são `https://certguard.franklem.uk`; routers, middleware e service Traefik usam prefixo `certguard-prod`, sem reutilizar os da homologação. Não clonar o banco, sessões, administrador ou segredos de homologação. Auto-deploy deve permanecer desligado, revisão fixada após CI e aceite.

A validação `scripts/check-coolify-compose.py <projeto> producao` confere isolamento, flags desativadas, roteamento e origem de produção. A CI renderiza ambos os arquivos com nomes padrão e UUIDs simulados. A configuração nova não autoriza a implantação antes de concluir backup externo/restauração e aceite da homologação.


## Backup externo do CertGuard

[Plano](plano-backup-externo.md), [operação](backup-externo.md), [ADR-012](adr/012-backup-externo.md) e [evidências/pendências](validation-backup-externo.md). Destino exclusivo confirmado em /srv/storage/backups/certguard no frank-note2; acesso SSH existente reutilizado. Scripts e testes preparados. Primeiro backup real, restauração, guarda da chave e agendamento ainda pendentes de entrada segura do responsável/acesso temporário. Não considerar esse gate aprovado apenas por existir a pasta.
