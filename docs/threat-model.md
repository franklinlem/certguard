# Modelo de ameaças — etapa inicial

| Ameaça | Ativo e caminho | Mitigação atual | Risco residual / próximo passo |
|---|---|---|---|
| SSRF / rebinding | Rede interna via hostname cadastrado | Todos os IPs validados, dial literal, porta 443 | Rotas/NAT exigem egress no host |
| Certificado inválido aceito | Confiança inferida de handshake | x509 + hostname obrigatórios | Revogação OCSP/CRL não implementada |
| DNS / TLS lento | CPU e sockets do worker | Timeout total e deadline | Pool limitado antes do scheduler |
| Metadados maliciosos | XSS e logs | JSON, frontend sem HTML dinâmico | Limites de armazenamento no próximo incremento |
| Brute force / sessão / CSRF | API administrativa | Sessão, CSRF, Origin e limites de login | Limites compartilhados e proxy confiável antes de replicar |
| SQL injection | Banco | SQL fixo, parâmetros pgx | Manter parâmetros no CRUD |
| Vazamento de secrets | Logs e repositório | Mensagens sanitizadas, .gitignore | Secret scanning e rotação no provisionamento |
| Storage exhaustion | Histórico infinito | Histórico de checks manuais | Retenção e quotas com scheduler |
| Alert storm / webhook SSRF | Saída de mensagens | Dedupe, limite de canais/tentativas, HMAC, TLS e DNS/IP público fixado | Monitorar volume e validar receptor na homologação |
| Supply chain | Imagens, dependências, CI | Lockfiles, testes, auditorias | Renovar imagens e revisar ações/dependências |

Etapa 06: chave perdida ou trocada pode impedir decriptação; preservar banco/chave separadamente e validar restauração. Queda após aceite pode repetir mensagens; receptor webhook deve deduplicar ID. SMTP não garante exatamente uma entrega. Chave não protege um processo já comprometido com acesso ao segredo; isolar app/segredos e limitar acesso administrativo.
