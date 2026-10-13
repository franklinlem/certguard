# Etapa 08 — homologação isolada

Data: 13/10/2026. Escopo autorizado: continuar a configuração da homologação após integração/inventário.

## Objetivo

Preparar recurso exclusivo CertGuard no Coolify da Contabo, com domínio `certguard-dev.franklem.uk`, HTTPS pelo Traefik existente e banco novo. Produção `certguard.franklem.uk` permanece sem implantação.

## Execução e verificações

1. Confirmar estado do repositório e CI do candidato e7970ba1 (migrações 1–6), capacidade disponível e recursos já existentes.
2. Criar projeto CertGuard/ambiente de homologação e aplicação Compose a partir do repositório privado, usando acesso GitHub já configurado; não ampliar permissões da integração.
3. Fixar revisão explícita, auto-deploy desligado, Compose base em deploy/compose.yaml e diretório/contexto de build corretos. Revisar Compose gerado: somente web:8080 no proxy, db/API sem portas públicas, redes/volume exclusivos.
4. Configurar APP_ENV=production, PUBLIC_APP_URL HTTPS exata, senha PostgreSQL exclusiva armazenada somente no ambiente. Scheduler/retenção/entregas desabilitados; métricas/chave de canais vazias nesta subida inicial.
5. Executar build/implantação apenas do recurso novo se o isolamento e orçamento forem preservados. Sem instalação/atualização do proxy ou limpeza de outros projetos.
6. Verificar migrações, health/readiness, HTTPS, interface, bloqueio de rotas privadas sem sessão e ausência de exposição de métricas. Registrar recursos e estado dos serviços existentes.
7. Bootstrap administrativo depende da identidade do responsável e entrada segura de credencial; não versionar nem enviar senha na conversa. Validação completa de monitoramento fica pendente até acesso administrativo e destino autorizado.

## Backup informado e limites

Destino escolhido pelo responsável: `franklin@frank-note2`, servidor físico Linux Mint XFCE, Docker, agente Hermes, outros backups e Uptime Kuma (informações fornecidas, ainda não verificadas). Verificação e configuração do backup ficam para etapa posterior, conforme solicitado. Não acessar ou modificar esse servidor neste incremento. Não usar dados reais ou ativar retenção/entregas antes de validar backup/restauração.

## Impacto e recuperação

Criação de recurso novo e volume vazio; orçamento de runtime db 512MiB/API 256MiB/web 128MiB, com build/migração adicionais a medir. Se a operação exigir mudar proxy global/firewall/rede de terceiros, parar e apresentar impacto/rollback para autorização específica. Em falha, parar apenas aplicação nova preservando volume e diagnóstico. Nunca prune, down -v, DROP ou downgrade automático de schema. Produção não faz parte desta autorização.

## Critério de conclusão

Configuração registrada e revisável, implantação isolada e verificações iniciais comprovadas ou bloqueio concreto documentado. HTTPS válido não é aceite completo: backup, bootstrap e fluxos autenticados devem constar como pendências quando ainda não executados.

## Ajuste após inspeção do Compose gerado

O modo gerenciado do Coolify 4.3.23 adicionou sua rede de proxy a todos os serviços, inclusive db/migrate/API. Para preservar isolamento, preparar compose.coolify.yaml na raiz com redes explícitas e labels somente no web, usando Raw (deploy file as-is). Rede externa coolify já confirmada no inventário. Contexto de build na raiz, nomes e domínio exclusivos da homologação. Validar política de isolamento e sintaxe na CI antes de selecionar o novo SHA. Não alterar configuração global do proxy.
