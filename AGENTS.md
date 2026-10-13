# Instruções de desenvolvimento do CertGuard

- Comunicar e escrever interfaces em português do Brasil.
- Antes de implementar um novo incremento, registrar planejamento em docs/; manter ADRs e evidências de validação.
- O arquivo docs/planejamento-completo.md descreve o produto; README e docs/validation.md identificam o escopo realmente implementado.
- Stack: Go, React/TypeScript, PostgreSQL, Docker Compose, GitHub Actions; implantação planejada em Coolify/Traefik.
- Não expor checks/CRUD sem autenticação, autorização e limites de uso. Não permitir bypass SSRF por configuração para facilitar testes.
- Testar o motor com DNS e servidores TLS isolados. Nunca monitorar destinos sem autorização.
- Homologação: certguard-dev.franklem.uk. Produção: certguard.franklem.uk. Bancos, volumes, redes e segredos separados.
- Não alterar outros serviços da VPS. Produção exige homologação validada. Pedir autorização antes de alterações de infraestrutura com risco de interrupção ou perda de dados.
- Não versionar segredos. Não executar limpeza global de Docker, down -v ou restauração destrutiva em ambientes reais.
- Rodar testes, lint e build proporcionais à mudança; CI deve permanecer verde.
