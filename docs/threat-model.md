# Modelo de ameaças — etapa inicial

| Ameaça | Ativo e caminho | Mitigação atual | Risco residual / próximo passo |
|---|---|---|---|
| SSRF / rebinding | Rede interna via hostname cadastrado | Todos os IPs validados, dial literal, porta 443 | Rotas/NAT exigem egress no host |
| Certificado inválido aceito | Confiança inferida de handshake | x509 + hostname obrigatórios | Revogação OCSP/CRL não implementada |
| DNS / TLS lento | CPU e sockets do worker | Timeout total e deadline | Pool limitado antes do scheduler |
| Metadados maliciosos | XSS e logs | JSON, frontend sem HTML dinâmico | Limites de armazenamento no próximo incremento |
| Brute force / sessão / CSRF | API administrativa | API de negócio ainda indisponível | Auth e testes antes da exposição |
| SQL injection | Banco | SQL fixo, parâmetros pgx | Manter parâmetros no CRUD |
| Vazamento de secrets | Logs e repositório | Mensagens sanitizadas, .gitignore | Secret scanning e rotação no provisionamento |
| Storage exhaustion | Histórico infinito | Sem persistência automática ainda | Retenção e quotas com scheduler |
| Alert storm / webhook SSRF | Saída de mensagens | Alertas ainda indisponíveis | Dedupe, assinatura e política SSRF própria |
| Supply chain | Imagens, dependências, CI | Lockfiles, testes, auditorias | Renovar imagens e revisar ações/dependências |
