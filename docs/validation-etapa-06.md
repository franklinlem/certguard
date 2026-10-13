# Validação da etapa 06

Base: eventos / PR #9. Sem merge, implantação, credenciais reais ou mensagens externas.

## Local

- Go: race, vet e staticcheck aprovados. PostgreSQL indisponível localmente; testes SQL aguardam CI.
- Transportes reais em fixtures locais: HTTPS assinado, Host/SNI, redirects, respostas transitórias/permanentes, TLS não confiável, DNS misto/privado, SMTP 465 e STARTTLS 587, recusa de autenticação sem TLS e Message-ID.
- Cifra: nonces distintos, AAD, adulteração, segredo ausente do ciphertext e validação de configuração.
- Frontend: tipos/lint/quatorze testes/build aprovados; confirmação de ativação, cadastro pausado com CSRF, remoção do segredo do formulário e indisponibilidade sem chave.

## CI

Aguardando PostgreSQL, Docker, Playwright e segurança. Testes SQL cobrem enfileiramento/dedupe, locks, cancelamento, retries, limite de canais, chave incompatível e rollback da verificação ao falhar a fila. E2E mantém executor desabilitado e não usa destinatários reais.

## Limites

Aceite não comprova chegada à caixa final. Possíveis duplicatas após queda antes do commit. Credenciais/destino imutáveis; sem retry manual, retenção de entregas/eventos ou rotação assistida de chave. Integração dos PRs e homologação continuam pendentes.
