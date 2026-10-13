# ADR-008 — canais cifrados e outbox de entregas

Status: proposto na etapa 06, dependente do ADR-007.

Canal global é um destino autorizado pelo administrador. Novo canal começa pausado; ativação exige mutação autenticada/CSRF e confirmação visual. Configuração é imutável para evitar reenviar pendências a outro destinatário após edição. Nome e ativação podem mudar; destino/credenciais novos exigem substituto. Limite de 20 canais não arquivados reduz fan-out de cada evento. Não há backfill ou envio de teste.

AES-256-GCM protege configuração em banco e na cópia de cada entrega, com nonce aleatório e AAD `channel_id:type`. A chave de 32 bytes vem de ALERT_ENCRYPTION_KEY/base64, nunca de valor padrão no código. Leituras retornam metadados, sem URL/segredo/senha/destinatário. Sem chave, consulta permanece disponível e cadastro/ativação ficam indisponíveis. Executor habilitado exige chave e verifica decriptação dos canais ativos antes de iniciar. Rotação assistida não está implementada: preservar chave e banco nos backups, sem trocar chave de instalação existente.

Cada evento cria uma entrega por canal ativo na mesma transação do check, incluindo snapshot cifrado da configuração e payload. FOR SHARE nos canais fecha a corrida com pausa/arquivo. Não há operação de envio dentro da transação do check. Evento/entrega não confirmam recepção externa até a execução posterior.

O processo serve executa um dispatcher por instância quando DELIVERY_ENABLED=true, usando pool separado da API, compartilhado com scheduler/retention (capacidade CHECK_WORKERS+2). Cada trabalho bloqueia entrega e canal com SKIP LOCKED até concluir transporte e salvar estado. Pausa/arquivo disputam o mesmo lock com NOWAIT e cancelam todas as pendências. Reativação não recria canceladas nem eventos antigos. Desabilitar o dispatcher preserva pendências; revisar backlog antes de reativar.

Timeout de transporte 10s, operação 15s. Até cinco tentativas, intervalos de 1/5/15/60 minutos e jitter até 10%. Estado terminal sent/failed/cancelled não é selecionado novamente. Registra contagem e último resultado, sem tabela individual de cada tentativa. Erros remotos são convertidos em códigos estáveis; nunca armazenar corpo de resposta ou erro bruto. Cancelamento do processo reverte a transação; queda após aceite e antes do commit pode repetir envio. X-CertGuard-Delivery e Message-ID estáveis permitem idempotência, mas SMTP não garante deduplicação.

Webhook exige HTTPS 443, TLS >=1.2 com confiança/hostname, DNS público integralmente validado por tentativa, dial em IP aprovado, sem proxy ou redirects. HMAC-SHA256 usa `timestamp + '.' + body` exatos. Corpo JSONB persistido fornece bytes estáveis entre tentativas; timestamp/assinatura renovam. Receptor deve validar assinatura, janela de tempo e ID de entrega.

SMTP permite apenas 465 com TLS implícito ou 587 com STARTTLS obrigatório. DNS/IP segue a política conservadora do motor, com porta SMTP explicitamente permitida. Credenciais são usadas somente após TLS validado. Uma mailbox ASCII simples por canal, sem cabeçalhos livres. Mensagem texto UTF-8/base64 com headers controlados. DATA aceito confirma aceite pelo servidor; erro posterior de QUIT não força retry. Não significa chegada à caixa final.

HTTP 2xx é aceite; 408/429/5xx são transitórios. Demais respostas, inclusive redirects, são permanentes. SMTP 4xx é transitório e 5xx permanente; falhas DNS/rede são transitórias, SSRF/config/TLS inválido permanentes. Limitações: sem edição de destino, retries manuais, políticas por endpoint, retenção de eventos/entregas ou rotação assistida nesta entrega. Implantação permanece sujeita à homologação e aos limites de infraestrutura do usuário.

Referências: [net/smtp](https://pkg.go.dev/net/smtp), [crypto/cipher](https://pkg.go.dev/crypto/cipher#NewGCM), [net/http Transport](https://pkg.go.dev/net/http#Transport).
