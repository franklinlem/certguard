# Validação da etapa 06

Base: eventos / PR #9. Sem merge, implantação, credenciais reais ou mensagens externas.

## Local

- Go: race, vet e staticcheck aprovados. PostgreSQL indisponível localmente; testes SQL aprovados na CI.
- Transportes reais em fixtures locais: HTTPS assinado, Host/SNI, redirects, respostas transitórias/permanentes, TLS não confiável, DNS misto/privado, SMTP 465 e STARTTLS 587, recusa de autenticação sem TLS e Message-ID.
- Cifra: nonces distintos, AAD, adulteração, segredo ausente do ciphertext e validação de configuração.
- Frontend: tipos/lint/quatorze testes/build aprovados; confirmação de ativação, cadastro pausado com CSRF, remoção do segredo do formulário e indisponibilidade sem chave.

## CI

Código validado: `5a5a3a87384e3bfbefe64e221dfd425eea8cfccd`.

- [Validação #38399819572](https://github.com/franklinlem/certguard/actions/runs/38399819572): PostgreSQL/race/vet/staticcheck, cobertura/fuzzing, frontend, Docker e Playwright aprovados.
- [Segurança #38399819552](https://github.com/franklinlem/certguard/actions/runs/38399819552): dependências, código e imagens aprovados.
- [PR #10](https://github.com/franklinlem/certguard/pull/10), dependente do PR #9, aberto sem merge.

Testes SQL confirmaram enfileiramento/dedupe, locks, cancelamento, retries, limite de canais, chave incompatível e rollback da verificação ao falhar a fila. E2E manteve executor desabilitado e chave ausente, validando leitura de canais/entregas e indisponibilidade de cadastro; formulários/ativação foram testados no frontend e API com cifra de teste. Transportes foram exercitados em servidores locais, não em um provedor real.

O commit posterior apenas registra evidências; código permanece igual ao validado. Homologação com destinatário/receptor autorizado e configuração real ainda é necessária antes de habilitar envios.

## Limites

Aceite não comprova chegada à caixa final. Possíveis duplicatas após queda antes do commit. Credenciais/destino imutáveis; sem retry manual, retenção de entregas/eventos ou rotação assistida de chave. Integração dos PRs e homologação continuam pendentes.
