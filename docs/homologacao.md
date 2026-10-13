# Roteiro de homologação — CertGuard

Estado: inventário somente leitura executado em 13/10/2026; implantação inicial executada; aceite completo ainda pendente. Consulte [evidências atuais](validation-etapa-08.md). Consulte [inventário](inventario-homologacao-2026-10-13.md) e [integração](integracao-prs-2026-10-13.md). Não é autorização para modificar recursos compartilhados.

## 1. Base revisável

- Integrar em ordem os PRs pendentes após revisão; reexecutar CI nas bases finais. Dependabot precisa ser avaliado no estado integrado.
- Fixar commit/imagens e registrar: SHA, migrações esperadas (1–6), execução CI, domínio `certguard-dev.franklem.uk`, responsável e data.
- Produção `certguard.franklem.uk` permanece separada e depende do aceite completo de homologação.

## 2. Inventário somente leitura da VPS

Registrar versões Docker/Compose/Coolify/Traefik, serviços existentes, CPU/memória/disco disponíveis, redes/portas, DNS/Cloudflare e backups. Confirmar onde o proxy conecta a aplicação e os nomes reais de rede/entrypoint/certresolver. Não exportar segredos de inspeções completas.

| Item | Evidência a preencher |
|---|---|
| Host/acesso e responsável | Contabo 158.220.120.2, terminal autenticado Coolify; Franklin |
| Recursos livres e margem para build/migração | 4974 MiB disponíveis e 78G livres; pico de build ainda não medido |
| Serviços/volumes/redes existentes a preservar | Inventariados; 25 containers existentes, nenhuma instância CertGuard |
| Rede e integração nativa do proxy | Traefik v3.7, http/https, letsencrypt; futura rede do recurso ainda não criada |
| DNS e HTTPS do domínio de homologação | A para VPS confirmado; HTTPS confiável Let's Encrypt verificado em 13/10/2026 |
| Commit/imagens e CI | Código 66db5902 implantado; CI aprovada; imagens construídas no Coolify |
| Destino de backup fora da VPS | franklin@frank-note2 escolhido; configuração/verificação adiada e ainda pendente |

Qualquer mudança com risco de interrupção/perda de dados exige autorização específica após apresentar impacto e rollback concretos. Não instalar segundo Traefik nem alterar firewall/redes/volumes de outros serviços por conveniência.

## 3. Aplicação isolada

Projeto `certguard-dev`, banco/volume/segredos próprios. Compose específico `/compose.coolify.yaml` em modo Raw nesta homologação, conforme ADR-010; nunca compose.dev na VPS. Somente web:8080 conectado pelo Coolify ao proxy existente. API/db sem portas públicas; sem docker.sock ou volumes de outros projetos. Auto-deploy desligado até validar rollback. Orçamento configurado inclui db 512MiB/API 256MiB/web 128MiB, além de build/migrate e serviços existentes; conferir disponibilidade real antes de executar.

Secrets: senha PostgreSQL alfanumérica forte, PUBLIC_APP_URL HTTPS exata, APP_ENV=production. Chave de canais e token de métricas opcionais nesta primeira subida; preservar fora do Git. `SCHEDULER_ENABLED=false`, `RETENTION_ENABLED=false`, `DELIVERY_ENABLED=false`. Readiness deve ficar 200 somente com schema completo.

## 4. Backup/restauração

Antes de dados reais, definir pg_dump custom do banco específico, criptografia/armazenamento externo e cópia separada da chave de canais. Restaurar em banco novo descartável com pg_restore sem --clean; conferir versões, contagens, integridade e decriptação com chave de teste. Nunca usar produção como destino de ensaio. O script scripts/ci-backup-restore.sh só aceita GitHub Actions e projeto efêmero; não é ferramenta de restauração da VPS.

Ensaio CI preserva schema/contagens/contadores, sem chave ou provedor reais. Testar restauração de chave/configuração no ambiente autorizado antes de habilitar entregas. Guardar evidência do teste e política de retenção dos backups.

## 5. Fluxos e aceite

- Bootstrap de administrador via stdin; login/logout, cookie Secure, CSRF, ausência de cache e isolamento de origens.
- Cadastro de destino autorizado, check manual, cadeia/hostname/SANs/validade/fingerprint e histórico. Testes maliciosos SSRF permanecem isolados, sem sondar serviços internos reais.
- Habilitar scheduler apenas para destino autorizado; verificar cadência, pause/retomada e reinício. Confirmar que outros serviços não mudaram.
- Cadastrar receptor/webhook ou mailbox de teste autorizados; revisar backlog e ativar canal antes do executor. Conferir assinatura, aceite/falha/retry/cancelamento. Um aceite SMTP não prova chegada à caixa final.
- Métricas por rede interna/token, sem acesso pelo domínio web. Registrar limites observados e política de coleta por banco.
- Desktop/mobile, estados vazio/erro e acessibilidade. Medir recursos e latências reais; números de CI não são capacidade da VPS.

## 6. Rollback e pendências

Guardar imagem/commit anterior compatível e procedimento de seleção no Coolify. Não apagar volumes ou reverter schema automaticamente. Readiness rejeita schema futuro desconhecido; avaliar compatibilidade antes de escolher binário anterior. Nunca down -v/prune/DROP/restauração destrutiva em servidor real sem plano/autorização.

Antes de uso prolongado/produção: definir retenção de eventos/entregas, rotação de chave, recuperação de acesso administrativo, limites de carga e observabilidade do ambiente. Nem o MVP estável nem homologação estão concluídos apenas com CI verde. Publicação em produção exige aceite documentado separado.
