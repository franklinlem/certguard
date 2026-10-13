# Etapa 06 — canais, SMTP, webhooks e fila de entregas

## Objetivo e base

Completar a entrega externa de alertas a partir da etapa 05 / PR #9, em branch e PR dependentes. Implementar canais globais, fila transacional, SMTP seguro, webhooks assinados e histórico de tentativas. Sem implantação, credenciais reais, envio real ou alteração de infraestrutura nesta execução.

## Canais e segredos

- Até 20 canais não arquivados por instalação; tipos webhook e smtp. Cadastro via API/painel autenticados, alteração de nome/ativação, arquivamento. Configuração do destino é imutável: criar substituto e arquivar anterior para trocar credenciais/destino.
- Configuração cifrada AES-256-GCM com nonce aleatório e AAD do UUID/tipo do canal. `ALERT_ENCRYPTION_KEY` em base64 (32 bytes), fora do banco/Git. Sem chave, cadastro indisponível; metadados/histórico continuam legíveis. API nunca devolve URL, senha, segredo HMAC ou destinatário.
- Cadastro não dispara entrega. Todos os canais novos começam pausados; ativação explícita requer CSRF e confirmação na interface. Configuração aceita apenas HTTPS 443 ou SMTP 465/587, DNS público e endereços de e-mail simples, sem cabeçalhos/CRLF.

## Eventos, fila e execução

- Ao criar evento, enfileirar uma entrega por canal ativo no mesmo commit do check. Snapshot da configuração cifrada e payload estável; sem backfill ou envio de teste nesta etapa.
- `DELIVERY_ENABLED=false` por padrão. Executor separado do pool da API, uma entrega por vez/instância, lock SKIP LOCKED mantido até concluir e persistir. Timeout total de 10s; operação de banco limitada. Encerramento cancela operações e aguarda workers.
- Até cinco tentativas; esperas de 1, 5, 15 e 60 minutos com jitter de até 10%. Sucesso terminal, falha permanente terminal; falhas transitórias recebem nova data. Cancelamento reverte tentativa. Queda após envio e antes do commit pode duplicar: IDs estáveis para idempotência e Message-ID SMTP estável.
- Pausar/arquivar canal cancela pendências e só afeta envios futuros/depois de concluir envio já em andamento. Reativação não reenvia histórico. Canais substituídos recebem somente eventos novos. Desabilitar o executor preserva fila, que será processada quando reativado; revisar backlog antes.

## Transportes

- Webhook: POST HTTPS, JSON de evento/observação, HMAC-SHA256 sobre `timestamp + '.' + body`; headers X-CertGuard-Event/Timestamp/Signature e ID estável da entrega. Não seguir redirects nem proxy de ambiente. DNS A/AAAA público revalidado a cada tentativa, conexão fixada em IP aprovado, hostname preservado para SNI e certificado validado.
- SMTP: TLS implícito 465 ou STARTTLS obrigatório 587, certificado/hostname verificados, autenticação somente após TLS, um destinatário por canal, mensagem de texto UTF-8, headers controlados. Mesma política de IP público do motor TLS; portas SMTP explicitamente restritas.
- Classificar falhas com códigos estáveis, sem registrar resposta remota, corpo, endereço, credenciais ou erro bruto. Webhook 2xx confirma recebimento; SMTP confirma aceite pelo servidor, não chegada à caixa final. Respostas permanentes 4xx (exceto 408/429), redirects e SMTP 5xx encerram; DNS/rede/5xx HTTP/4xx SMTP retentáveis, bloqueio SSRF/configuração/TLS inválido permanentes.

## API e interface

- GET/POST /api/v1/alert-channels; PATCH/DELETE /{id}, mutações com sessão/Origin/CSRF. Metadados sem configuração sensível. Leituras paginadas de entregas em GET /api/v1/alert-deliveries, filtro por evento opcional.
- Sessão informa disponibilidade da chave e estado do executor. Painel cadastra canais pausados, confirma ativação/arquivo, mostra estado da fila e tentativas. Mensagens distinguem aceitação do destino de entrega final.

## Validação e limites

- Testes de cifra/adulteração/AAD, validação de config e headers, assinatura, SSRF/rebinding/DNS misto, TLS obrigatório e redirects.
- Transportes reais em servidores TLS/SMTP isolados com resolver/dial/CA privados dos testes; nenhum bypass configurável por ambiente.
- PostgreSQL: enfileiramento atômico, dedupe, snapshot cifrado, concorrência, retries, cancelamento, pausa/arquivo, canais limitados e dados ausentes/chave incorreta.
- HTTP/React/E2E: autenticação, CSRF, segredos ausentes em respostas, cadastro sem enviar e operação indisponível sem chave. CI Go/PostgreSQL/Docker/navegador/segurança.
- Sem edição de destino, retries manuais, filtros por endpoint ou políticas por canal nesta entrega. Retenção de eventos/entregas e rotação assistida de chave ficam no hardening. Backups precisam preservar banco e chave separadamente. Produção exige homologação validada e autorização conforme risco.
