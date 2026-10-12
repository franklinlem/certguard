# Validação da etapa inicial

Em 12/10/2026:
- Go 1.27.2: build e go vet aprovados.
- go test -race: testes locais aprovados.
- Cobertura: certificate 91,8%; security 97,8%; httpapi 100%.
- Fuzzing do parser: 64.421 entradas em cerca de 11 segundos, sem falhas.
- Testes usam certificados efêmeros e servidores locais; nenhum domínio externo foi varrido.
- PostgreSQL local não instalado; teste de migração foi explicitamente pulado localmente. CI possui PostgreSQL 17 e testa idempotência, constraints e gravação/leitura de histórico.
- Docker não está instalado neste computador. Build e smoke test Compose estão configurados em runner efêmero do GitHub, sem acesso à VPS.

Resultados da CI e validação web serão registrados ao concluir a publicação. Nenhuma implantação em homologação ou produção foi executada.

- Staticcheck v0.8.2: aprovado.
- Frontend: lint, TypeScript, teste de componente, build e npm audit aprovados.
- Auditoria Go detectou vulnerabilidades transitivas em x/text; a dependência foi atualizada para v0.41.0 e a suíte passou novamente. Govulncheck v1.9.0: nenhuma vulnerabilidade encontrada.
