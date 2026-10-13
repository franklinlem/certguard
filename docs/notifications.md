# Operação de canais e entregas

Esta funcionalidade está implementada para validação em homologação; nenhuma mensagem real foi enviada durante o desenvolvimento.

## Configuração

`ALERT_ENCRYPTION_KEY` deve conter 32 bytes aleatórios codificados em base64. Gere no ambiente autorizado e salve como segredo do CertGuard no Coolify; preserve cópia segura separada do backup do banco. Não cole a chave em chats, argumentos públicos ou arquivos versionados. Sem chave, canais novos não podem ser cadastrados/ativados. Não troque a chave de uma instalação existente sem migração: pode tornar configurações e pendências indecifráveis. O executor verifica a chave dos canais ativos ao iniciar.

`DELIVERY_ENABLED=false` é o padrão. Habilitar é decisão operacional explícita, somente após revisar canais, destinos autorizados e backlog. Canais ativos enfileiram eventos mesmo com executor desabilitado. A reativação do executor processará pendências antigas. Não habilitar produção antes de homologar; não alterar proxy/serviços compartilhados da VPS.

No painel **Canais e entregas**, cadastre um canal pausado. Webhook pede URL HTTPS/443 e segredo de assinatura de pelo menos 32 caracteres. SMTP pede hostname público, porta 465 ou 587, usuário/senha, remetente e um destinatário simples. Confirme ativação quando o destino estiver revisado. Cadastro e ativação não enviam teste nem histórico anterior; somente eventos novos são enfileirados.

Para mudar destino ou credenciais, crie outro canal e arquive o anterior. Pausar/arquivar cancela pendências de forma terminal; reativar não as recria. Se houver envio em andamento, a alteração retorna conflito e deve ser tentada após a conclusão. Preservar histórico significa que canal arquivado e suas configurações cifradas continuam no banco.

## Webhook

POST JSON com event_id, event, severity, timestamp, endpoint, code e certificado/dias quando disponíveis. O corpo não contém credenciais do canal. Headers:

- X-CertGuard-Event: tipo do evento.
- X-CertGuard-Timestamp: Unix seconds da tentativa.
- X-CertGuard-Signature: `sha256=<hex HMAC-SHA256(secret, timestamp + '.' + body)>`.
- X-CertGuard-Delivery: UUID estável da entrega, chave de idempotência recomendada.

Verificar assinatura com comparação constante, tolerância de tempo (por exemplo, cinco minutos) e deduplicar pelo ID da entrega. Respostas 2xx confirmam aceite. Redirects não são seguidos. DNS é revalidado a cada tentativa, inclusive registros IPv6; mistura de IP público e privado bloqueia a saída. Não há opção de desabilitar TLS ou SSRF.

## SMTP

TLS obrigatório: 465 implícito ou 587 STARTTLS. Confiança e hostname são validados antes de autenticar. Cada canal tem um destinatário; múltiplos destinatários exigem canais separados. SMTP DATA aceito significa aceite pelo servidor, sujeito a antispam, bounce e políticas do provedor. Message-ID estável ajuda a reconhecer duplicatas, sem garantir deduplicação. DKIM/SPF/DMARC pertencem ao provedor/domínio remetente e devem ser revisados na homologação.

## Fila

Um trabalho por instância, timeout de 10s, no máximo cinco tentativas. Falhas transitórias aguardam 1/5/15/60 minutos com jitter até 10%. Erros permanentes terminam imediatamente. Histórico mostra contagem e último código, sem guardar respostas remotas ou uma linha por tentativa. `sent` aparece como **Aceita pelo destino**, sem prometer entrega à caixa de e-mail.

Uma queda depois do aceite remoto e antes do commit pode duplicar envio. Nenhum mecanismo SMTP garante exatamente uma entrega. Durante homologação, use apenas destinatário/receptor de teste autorizado, confirme estados pending/sent/failed/cancelled, pausa, reinício e validade da assinatura. Testes automatizados usam servidores locais e não enviam mensagens públicas.

Retenção de eventos e entregas, rotação assistida da chave e retry manual ficam para hardening. Inclua banco e chave no procedimento de backup/restauração antes de uso real.
